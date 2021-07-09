/*
Copyright 2021 The k8gb Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Generated by GoLic, for more details see: https://github.com/AbsaOSS/golic
*/
package test

import (
	"k8gbterratest/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFullRoundRobin(t *testing.T) {
	t.Parallel()
	const host = "roundrobin-test.cloud.example.com"
	const gslbPath = "../examples/roundrobin2.yaml"

	instanceEU, err := utils.NewWorkflow(t, "k3d-test-gslb1", 5053).
		WithGslb(gslbPath, host).
		WithTestApp("eu").
		Start()
	require.NoError(t, err)
	defer instanceEU.Kill()
	instanceUS, err := utils.NewWorkflow(t, "k3d-test-gslb2", 5054).
		WithGslb(gslbPath, host).
		WithTestApp("us").
		Start()
	require.NoError(t, err)
	defer instanceUS.Kill()

	t.Run("round-robin on two concurrent clusters with podinfo running", func(t *testing.T) {
		err = instanceEU.WaitForAppIsRunning()
		require.NoError(t, err)
		err = instanceUS.WaitForAppIsRunning()
		require.NoError(t, err)
	})

	euLocalTargets := instanceEU.GetLocalTargets()
	usLocalTargets := instanceUS.GetLocalTargets()
	expectedIPs := append(euLocalTargets, usLocalTargets...)

	t.Run("kill podinfo on the second cluster", func(t *testing.T) {
		instanceUS.StopTestApp()
		err = instanceEU.WaitForExpected(euLocalTargets)
		require.NoError(t, err)
		err = instanceUS.WaitForExpected(euLocalTargets)
		require.NoError(t, err)
	})

	t.Run("kill podinfo on the first cluster", func(t *testing.T) {
		instanceEU.StopTestApp()
		err = instanceUS.WaitForExpected([]string{})
		require.NoError(t, err)
		err = instanceEU.WaitForExpected([]string{})
		require.NoError(t, err)
	})

	t.Run("start podinfo on the second cluster", func(t *testing.T) {
		instanceUS.StartTestApp()
		err = instanceEU.WaitForExpected(usLocalTargets)
		require.NoError(t, err)
		err = instanceUS.WaitForExpected(usLocalTargets)
		require.NoError(t, err)
	})

	t.Run("start podinfo on the first cluster", func(t *testing.T) {
		// start app in the both clusters
		instanceEU.StartTestApp()
		err = instanceEU.WaitForExpected(expectedIPs)
		require.NoError(t, err)
		err = instanceUS.WaitForExpected(expectedIPs)
		require.NoError(t, err)
	})
}
