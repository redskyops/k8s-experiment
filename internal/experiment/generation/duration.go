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
	redskyappsv1alpha1 "github.com/thestormforge/optimize-controller/api/apps/v1alpha1"
	redskyv1beta1 "github.com/thestormforge/optimize-controller/api/v1beta1"
)

type DurationMetricsSource struct {
	Goal *redskyappsv1alpha1.Goal
}

var _ MetricSource = &DurationMetricsSource{}

func (s *DurationMetricsSource) Metrics() ([]redskyv1beta1.Metric, error) {
	var result []redskyv1beta1.Metric
	if s.Goal == nil || s.Goal.Implemented {
		return result, nil
	}

	switch s.Goal.Duration.DurationType {
	case redskyappsv1alpha1.DurationTrial:
		m := newGoalMetric(s.Goal, `{{ duration .StartTime .CompletionTime }}`)
		m.Type = ""
		result = append(result, m)
	}

	return result, nil
}
