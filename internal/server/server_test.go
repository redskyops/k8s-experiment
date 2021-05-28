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

package server

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	optimizev1beta1 "github.com/thestormforge/optimize-controller/v2/api/v1beta1"
	experimentsv1alpha1 "github.com/thestormforge/optimize-go/pkg/api/experiments/v1alpha1"
	"github.com/thestormforge/optimize-go/pkg/api/experiments/v1alpha1/numstr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestFromCluster(t *testing.T) {
	one := intstr.FromInt(1)
	two := intstr.FromInt(2)
	three := intstr.FromString("three")
	now := time.Now()
	cases := []struct {
		desc     string
		in       *optimizev1beta1.Experiment
		out      *experimentsv1alpha1.Experiment
		name     experimentsv1alpha1.ExperimentName
		baseline *experimentsv1alpha1.TrialAssignments
	}{
		{
			desc: "basic",
			in: &optimizev1beta1.Experiment{
				ObjectMeta: metav1.ObjectMeta{
					Name: "basic",
					CreationTimestamp: metav1.Time{
						Time: now,
					},
					Annotations: map[string]string{
						optimizev1beta1.AnnotationExperimentURL: "self_111",
						optimizev1beta1.AnnotationNextTrialURL:  "next_trial_111",
					},
				},
			},
			out: &experimentsv1alpha1.Experiment{
				ExperimentMeta: experimentsv1alpha1.ExperimentMeta{
					LastModified: now,
					SelfURL:      "self_111",
					NextTrialURL: "next_trial_111",
				},
			},
		},
		{
			desc: "optimization",
			in: &optimizev1beta1.Experiment{
				Spec: optimizev1beta1.ExperimentSpec{
					Optimization: []optimizev1beta1.Optimization{
						{Name: "one", Value: "111"},
						{Name: "two", Value: "222"},
						{Name: "three", Value: "333"},
					},
				},
			},
			out: &experimentsv1alpha1.Experiment{
				Optimization: []experimentsv1alpha1.Optimization{
					{Name: "one", Value: "111"},
					{Name: "two", Value: "222"},
					{Name: "three", Value: "333"},
				},
			},
		},
		{
			desc: "parameters",
			in: &optimizev1beta1.Experiment{
				Spec: optimizev1beta1.ExperimentSpec{
					Parameters: []optimizev1beta1.Parameter{
						{Name: "one", Min: 111, Max: 222},
						{Name: "two", Min: 1111, Max: 2222},
						{Name: "three", Min: 11111, Max: 22222},
						{Name: "test_case", Min: 1, Max: 1},
					},
				},
			},
			out: &experimentsv1alpha1.Experiment{
				Parameters: []experimentsv1alpha1.Parameter{
					{
						Type: experimentsv1alpha1.ParameterTypeInteger,
						Name: "one",
						Bounds: &experimentsv1alpha1.Bounds{
							Min: json.Number(strconv.FormatInt(111, 10)),
							Max: json.Number(strconv.FormatInt(222, 10)),
						},
					},
					{
						Type: experimentsv1alpha1.ParameterTypeInteger,
						Name: "two",
						Bounds: &experimentsv1alpha1.Bounds{
							Min: json.Number(strconv.FormatInt(1111, 10)),
							Max: json.Number(strconv.FormatInt(2222, 10)),
						},
					},
					{
						Type: experimentsv1alpha1.ParameterTypeInteger,
						Name: "three",
						Bounds: &experimentsv1alpha1.Bounds{
							Min: json.Number(strconv.FormatInt(11111, 10)),
							Max: json.Number(strconv.FormatInt(22222, 10)),
						},
					},
				},
			},
		},
		{
			desc: "orderConstraints",
			in: &optimizev1beta1.Experiment{
				Spec: optimizev1beta1.ExperimentSpec{
					Constraints: []optimizev1beta1.Constraint{
						{
							Name: "one-two",
							Order: &optimizev1beta1.OrderConstraint{
								LowerParameter: "one",
								UpperParameter: "two",
							},
						},
					},
				},
			},
			out: &experimentsv1alpha1.Experiment{
				Constraints: []experimentsv1alpha1.Constraint{
					{
						ConstraintType:  experimentsv1alpha1.ConstraintOrder,
						Name:            "one-two",
						OrderConstraint: experimentsv1alpha1.OrderConstraint{LowerParameter: "one", UpperParameter: "two"},
					},
				},
			},
		},
		{
			desc: "sumConstraints",
			in: &optimizev1beta1.Experiment{
				Spec: optimizev1beta1.ExperimentSpec{
					Constraints: []optimizev1beta1.Constraint{
						{
							Name: "one-two",
							Sum: &optimizev1beta1.SumConstraint{
								Bound: resource.MustParse("1"),
								Parameters: []optimizev1beta1.SumConstraintParameter{
									{
										Name:   "one",
										Weight: resource.MustParse("-1.0"),
									},
									{
										Name:   "two",
										Weight: resource.MustParse("1"),
									},
									{
										Name:   "three",
										Weight: resource.MustParse("3.5"),
									},
									{
										Name:   "four",
										Weight: resource.MustParse("5000m"),
									},
								},
							},
						},
					},
				},
			},
			out: &experimentsv1alpha1.Experiment{
				Constraints: []experimentsv1alpha1.Constraint{
					{
						Name:           "one-two",
						ConstraintType: experimentsv1alpha1.ConstraintSum,
						SumConstraint: experimentsv1alpha1.SumConstraint{
							Bound: 1,
							Parameters: []experimentsv1alpha1.SumConstraintParameter{
								{Name: "one", Weight: -1.0},
								{Name: "two", Weight: 1.0},
								{Name: "three", Weight: 3.5},
								{Name: "four", Weight: 5.0},
							},
						},
					},
				},
			},
		},
		{
			desc: "metrics",
			in: &optimizev1beta1.Experiment{
				Spec: optimizev1beta1.ExperimentSpec{
					Metrics: []optimizev1beta1.Metric{
						{Name: "one", Minimize: true},
						{Name: "two", Minimize: false},
						{Name: "three", Minimize: true},
					},
				},
			},
			out: &experimentsv1alpha1.Experiment{
				Metrics: []experimentsv1alpha1.Metric{
					{Name: "one", Minimize: true},
					{Name: "two", Minimize: false},
					{Name: "three", Minimize: true},
				},
			},
		},
		{
			desc: "baseline",
			in: &optimizev1beta1.Experiment{
				Spec: optimizev1beta1.ExperimentSpec{
					Parameters: []optimizev1beta1.Parameter{
						{Name: "one", Min: 0, Max: 1, Baseline: &one},
						{Name: "two", Min: 0, Max: 2, Baseline: &two},
						{Name: "three", Values: []string{"three"}, Baseline: &three},
					},
				},
			},
			out: &experimentsv1alpha1.Experiment{
				Parameters: []experimentsv1alpha1.Parameter{
					{
						Type:   experimentsv1alpha1.ParameterTypeInteger,
						Name:   "one",
						Bounds: &experimentsv1alpha1.Bounds{Min: "0", Max: "1"},
					},
					{
						Type:   experimentsv1alpha1.ParameterTypeInteger,
						Name:   "two",
						Bounds: &experimentsv1alpha1.Bounds{Min: "0", Max: "2"},
					},
					{
						Type:   experimentsv1alpha1.ParameterTypeCategorical,
						Name:   "three",
						Values: []string{"three"},
					},
				},
			},
			baseline: &experimentsv1alpha1.TrialAssignments{
				Labels: map[string]string{"baseline": "true"},
				Assignments: []experimentsv1alpha1.Assignment{
					{ParameterName: "one", Value: numstr.FromInt64(1)},
					{ParameterName: "two", Value: numstr.FromInt64(2)},
					{ParameterName: "three", Value: numstr.FromString("three")},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			name, out, baseline, err := FromCluster(c.in)
			if assert.NoError(t, err) {
				assert.Equal(t, c.in.Name, name.Name())
				assert.Equal(t, c.out, out)
				assert.Equal(t, c.baseline, baseline)
			}
		})
	}
}

func TestToCluster(t *testing.T) {
	cases := []struct {
		desc   string
		exp    *optimizev1beta1.Experiment
		ee     *experimentsv1alpha1.Experiment
		expOut *optimizev1beta1.Experiment
	}{
		{
			desc: "basic",
			exp: &optimizev1beta1.Experiment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: nil,
				},
			},
			ee: &experimentsv1alpha1.Experiment{
				ExperimentMeta: experimentsv1alpha1.ExperimentMeta{
					SelfURL:      "self_111",
					NextTrialURL: "next_trial_111",
				},
				Optimization: []experimentsv1alpha1.Optimization{
					{Name: "one", Value: "111"},
					{Name: "two", Value: "222"},
					{Name: "three", Value: "333"},
				},
			},
			expOut: &optimizev1beta1.Experiment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						optimizev1beta1.AnnotationExperimentURL: "self_111",
						optimizev1beta1.AnnotationNextTrialURL:  "next_trial_111",
					},
					Finalizers: []string{
						Finalizer,
					},
				},
				Spec: optimizev1beta1.ExperimentSpec{
					Optimization: []optimizev1beta1.Optimization{
						{Name: "one", Value: "111"},
						{Name: "two", Value: "222"},
						{Name: "three", Value: "333"},
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			ToCluster(c.exp, c.ee)
			assert.Equal(t, c.expOut, c.exp)
		})
	}
}

func TestToClusterTrial(t *testing.T) {
	cases := []struct {
		desc       string
		trial      *optimizev1beta1.Trial
		suggestion *experimentsv1alpha1.TrialAssignments
		trialOut   *optimizev1beta1.Trial
	}{
		{
			desc: "empty name with generate name",
			trial: &optimizev1beta1.Trial{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "generate_name",
					Annotations:  map[string]string{},
				},
			},
			suggestion: &experimentsv1alpha1.TrialAssignments{
				TrialMeta: experimentsv1alpha1.TrialMeta{
					SelfURL: "some/path/1",
				},
				Assignments: []experimentsv1alpha1.Assignment{
					{ParameterName: "one", Value: numstr.FromInt64(111)},
					{ParameterName: "two", Value: numstr.FromInt64(222)},
					{ParameterName: "three", Value: numstr.FromInt64(333)},
				},
			},
			trialOut: &optimizev1beta1.Trial{
				ObjectMeta: metav1.ObjectMeta{
					Name:         "generate_name001",
					GenerateName: "generate_name",
					Annotations: map[string]string{
						optimizev1beta1.AnnotationReportTrialURL: "some/path/1",
					},
					Finalizers: []string{
						Finalizer,
					},
				},
				Status: optimizev1beta1.TrialStatus{
					Phase:       "Created",
					Assignments: "one=111, two=222, three=333",
				},
				Spec: optimizev1beta1.TrialSpec{
					Assignments: []optimizev1beta1.Assignment{
						{Name: "one", Value: intstr.FromInt(111)},
						{Name: "two", Value: intstr.FromInt(222)},
						{Name: "three", Value: intstr.FromInt(333)},
					},
				},
			},
		},
		{
			desc: "name with generate name",
			trial: &optimizev1beta1.Trial{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "generate_name",
					Annotations:  map[string]string{},
				},
			},
			suggestion: &experimentsv1alpha1.TrialAssignments{
				TrialMeta: experimentsv1alpha1.TrialMeta{
					SelfURL: "some/path/one",
				},
				Assignments: []experimentsv1alpha1.Assignment{
					{ParameterName: "one", Value: numstr.FromInt64(111)},
					{ParameterName: "two", Value: numstr.FromInt64(222)},
					{ParameterName: "three", Value: numstr.FromInt64(333)},
				},
			},
			trialOut: &optimizev1beta1.Trial{
				ObjectMeta: metav1.ObjectMeta{
					Name:         "generate_nameone",
					GenerateName: "generate_name",
					Annotations: map[string]string{
						optimizev1beta1.AnnotationReportTrialURL: "some/path/one",
					},
					Finalizers: []string{
						Finalizer,
					},
				},
				Status: optimizev1beta1.TrialStatus{
					Phase:       "Created",
					Assignments: "one=111, two=222, three=333",
				},
				Spec: optimizev1beta1.TrialSpec{
					Assignments: []optimizev1beta1.Assignment{
						{Name: "one", Value: intstr.FromInt(111)},
						{Name: "two", Value: intstr.FromInt(222)},
						{Name: "three", Value: intstr.FromInt(333)},
					},
				},
			},
		},
		{
			desc: "32bit overflow",
			trial: &optimizev1beta1.Trial{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			suggestion: &experimentsv1alpha1.TrialAssignments{
				Assignments: []experimentsv1alpha1.Assignment{
					{ParameterName: "overflow", Value: numstr.FromInt64(math.MaxInt64)},
				},
			},
			trialOut: &optimizev1beta1.Trial{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"redskyops.dev/report-trial-url": "",
					},
					Finalizers: []string{"serverFinalizer.redskyops.dev"},
				},
				Spec: optimizev1beta1.TrialSpec{
					Assignments: []optimizev1beta1.Assignment{
						{Name: "overflow", Value: intstr.FromInt(math.MaxInt32)},
					},
				},
				Status: optimizev1beta1.TrialStatus{
					Phase:       "Created",
					Assignments: fmt.Sprintf("overflow=%d", math.MaxInt32),
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			ToClusterTrial(c.trial, c.suggestion)
			assert.Equal(t, c.trialOut, c.trial)
		})
	}
}

func TestFromClusterTrial(t *testing.T) {
	cases := []struct {
		desc        string
		experiment  optimizev1beta1.Experiment
		trial       optimizev1beta1.Trial
		expectedOut *experimentsv1alpha1.TrialValues
	}{
		{
			desc: "no conditions",
			trial: optimizev1beta1.Trial{
				Status: optimizev1beta1.TrialStatus{
					Conditions: []optimizev1beta1.TrialCondition{},
				},
			},
			expectedOut: &experimentsv1alpha1.TrialValues{},
		},
		{
			desc: "not failed",
			trial: optimizev1beta1.Trial{
				Status: optimizev1beta1.TrialStatus{
					Conditions: []optimizev1beta1.TrialCondition{
						{Type: optimizev1beta1.TrialComplete, Status: corev1.ConditionTrue},
					},
				},
			},
			expectedOut: &experimentsv1alpha1.TrialValues{},
		},
		{
			desc: "failed",
			trial: optimizev1beta1.Trial{
				Status: optimizev1beta1.TrialStatus{
					Conditions: []optimizev1beta1.TrialCondition{
						{
							Type:    optimizev1beta1.TrialFailed,
							Status:  corev1.ConditionTrue,
							Reason:  corev1.PodReasonUnschedulable,
							Message: "0/3 nodes are available: 3 Insufficient cpu.",
						},
					},
				},
			},
			expectedOut: &experimentsv1alpha1.TrialValues{
				Failed:         true,
				FailureReason:  corev1.PodReasonUnschedulable,
				FailureMessage: "0/3 nodes are available: 3 Insufficient cpu.",
			},
		},
		{
			desc: "conditions not failed",
			trial: optimizev1beta1.Trial{
				Status: optimizev1beta1.TrialStatus{
					Conditions: []optimizev1beta1.TrialCondition{
						{Type: optimizev1beta1.TrialComplete, Status: corev1.ConditionTrue},
					},
				},
				Spec: optimizev1beta1.TrialSpec{
					Values: []optimizev1beta1.Value{
						{Name: "one", Value: "111.111", Error: "1111.1111"},
						{Name: "two", Value: "222.222", Error: "2222.2222"},
						{Name: "three", Value: "333.333", Error: "3333.3333"},
					},
				},
			},
			expectedOut: &experimentsv1alpha1.TrialValues{
				Values: []experimentsv1alpha1.Value{
					{MetricName: "one", Value: 111.111, Error: 1111.1111},
					{MetricName: "two", Value: 222.222, Error: 2222.2222},
					{MetricName: "three", Value: 333.333, Error: 3333.3333},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			out := FromClusterTrial(&c.trial)
			assert.Equal(t, c.expectedOut, out)
		})
	}
}

func TestStopExperiment(t *testing.T) {
	cases := []struct {
		desc        string
		exp         *optimizev1beta1.Experiment
		err         error
		expectedOut bool
		expectedExp *optimizev1beta1.Experiment
	}{
		{
			desc: "no error",
			exp: &optimizev1beta1.Experiment{
				ObjectMeta: metav1.ObjectMeta{},
			},
			err:         nil,
			expectedOut: false,
			expectedExp: &optimizev1beta1.Experiment{
				ObjectMeta: metav1.ObjectMeta{},
			},
		},
		{
			desc: "error wrong type",
			exp: &optimizev1beta1.Experiment{
				ObjectMeta: metav1.ObjectMeta{},
			},
			err: &experimentsv1alpha1.Error{
				Type: experimentsv1alpha1.ErrExperimentNameInvalid,
			},
			expectedOut: false,
			expectedExp: &optimizev1beta1.Experiment{
				ObjectMeta: metav1.ObjectMeta{},
			},
		},
		{
			desc: "error",
			exp: &optimizev1beta1.Experiment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						optimizev1beta1.AnnotationNextTrialURL: "111",
					},
				},
			},
			err: &experimentsv1alpha1.Error{
				Type: experimentsv1alpha1.ErrExperimentStopped,
			},
			expectedOut: true,
			expectedExp: &optimizev1beta1.Experiment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			out := StopExperiment(c.exp, c.err)
			assert.Equal(t, c.expectedOut, out)
			assert.Equal(t, c.expectedExp.GetAnnotations(), c.exp.GetAnnotations())
		})
	}
}
