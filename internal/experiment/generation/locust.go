/*
Copyright 2020 GramLabs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package generation

import (
	"fmt"

	redskyappsv1alpha1 "github.com/thestormforge/optimize-controller/api/apps/v1alpha1"
	redskyv1beta1 "github.com/thestormforge/optimize-controller/api/v1beta1"
	"github.com/thestormforge/optimize-controller/internal/sfio"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type LocustSource struct {
	Scenario    *redskyappsv1alpha1.Scenario
	Objective   *redskyappsv1alpha1.Objective
	Application *redskyappsv1alpha1.Application
}

var _ ExperimentSource = &LocustSource{} // Update trial job
var _ MetricSource = &LocustSource{}     // Locust specific metrics
var _ kio.Reader = &LocustSource{}       // ConfigMap for the locustfile.py

func (s *LocustSource) Update(exp *redskyv1beta1.Experiment) error {
	if s.Scenario == nil || s.Application == nil {
		return nil
	}

	pod := &ensureTrialJobPod(exp).Spec
	pod.Containers = []corev1.Container{
		{
			Name:  "locust",
			Image: trialJobImage("locust"),
			Env:   s.locustEnv(),
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "locustfile",
					ReadOnly:  true,
					MountPath: "/mnt/locust",
				},
			},
		},
	}

	pod.Volumes = []corev1.Volume{
		{
			Name: "locustfile",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: s.locustConfigMapName(),
					},
				},
			},
		},
	}

	// TODO We need to rethink how ingress scanning works, this just preserves existing behavior
	var ingressURL string
	if s.Application != nil && s.Application.Ingress != nil {
		ingressURL = s.Application.Ingress.URL
	}
	if ingressURL == "" {
		return fmt.Errorf("ingress must be configured when using Locust scenarios")
	}
	pod.Containers[0].Env = append(pod.Containers[0].Env, corev1.EnvVar{Name: "HOST", Value: ingressURL})

	return nil
}

func (s *LocustSource) Read() ([]*yaml.RNode, error) {
	result := sfio.ObjectSlice{}

	if s.Scenario.Locust.Locustfile != "" {
		data, err := loadApplicationData(s.Application, s.Scenario.Locust.Locustfile)
		if err != nil {
			return nil, err
		}

		cm := &corev1.ConfigMap{}
		cm.Name = s.locustConfigMapName()
		cm.Data = map[string]string{"locustfile.py": string(data)}
		result = append(result, cm)
	} else {
		return nil, fmt.Errorf("missing Locust file for scenario %q", s.Scenario.Name)
	}

	return result.Read()
}

func (s *LocustSource) Metrics() ([]redskyv1beta1.Metric, error) {
	var result []redskyv1beta1.Metric
	if s.Objective == nil {
		return result, nil
	}

	for i := range s.Objective.Goals {
		goal := &s.Objective.Goals[i]
		switch {

		case goal.Implemented:
			// Do nothing

		case goal.Latency != nil:
			if l := s.locustLatency(goal.Latency.LatencyType); l != "" {
				query := `scalar(` + l + `{job="trialRun",instance="{{ .Trial.Name }}"})`
				result = append(result, newGoalMetric(goal, query))
			}

		case goal.ErrorRate != nil:
			if goal.ErrorRate.ErrorRateType == redskyappsv1alpha1.ErrorRateRequests {
				query := `scalar(failure_count{job="trialRun",instance="{{ .Trial.Name }}"} / request_count{job="trialRun",instance="{{ .Trial.Name }}"})`
				result = append(result, newGoalMetric(goal, query))
			}

		}
	}

	return result, nil
}

func (s *LocustSource) locustConfigMapName() string {
	return fmt.Sprintf("%s-locustfile", s.Scenario.Name)
}

func (s *LocustSource) locustEnv() []corev1.EnvVar {
	var env []corev1.EnvVar

	if users := s.Scenario.Locust.Users; users != nil {
		env = append(env, corev1.EnvVar{
			Name:  "NUM_USERS",
			Value: fmt.Sprintf("%d", *users),
		})
	}

	if spawnRate := s.Scenario.Locust.SpawnRate; spawnRate != nil {
		env = append(env, corev1.EnvVar{
			Name:  "SPAWN_RATE",
			Value: fmt.Sprintf("%d", *spawnRate),
		})
	}

	if runTime := s.Scenario.Locust.RunTime; runTime != nil {
		// Just give Locust the number of seconds
		// See: https://github.com/locustio/locust/blob/1f30d36d8f8d646eccb55aab7080fa69bf35c0d7/locust/util/timespan.py
		env = append(env, corev1.EnvVar{
			Name:  "RUN_TIME",
			Value: fmt.Sprintf("%.0f", runTime.Seconds()),
		})
	}

	return env
}

func (s *LocustSource) locustLatency(lt redskyappsv1alpha1.LatencyType) string {
	switch redskyappsv1alpha1.FixLatency(lt) {
	case redskyappsv1alpha1.LatencyMinimum:
		return "min_response_time"
	case redskyappsv1alpha1.LatencyMaximum:
		return "max_response_time"
	case redskyappsv1alpha1.LatencyMean:
		return "average_response_time"
	case redskyappsv1alpha1.LatencyPercentile50:
		return "p50"
	case redskyappsv1alpha1.LatencyPercentile95:
		return "p95"
	case redskyappsv1alpha1.LatencyPercentile99:
		return "p99"
	default:
		return ""
	}
}
