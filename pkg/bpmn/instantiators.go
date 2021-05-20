// Copyright (c) 2021 Aree Enterprises, Inc. and Contributors
// Use of this software is governed by the Business Source License
// included in the file LICENSE
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/LICENSE-Apache-2.0

package bpmn

// InstantiatingFlowNodes returns a list of flow nodes that
// can instantiate the process.
func (p *Process) InstantiatingFlowNodes() (result []FlowNodeInterface) {
	result = make([]FlowNodeInterface, 0)

	for i := range *p.StartEvents() {
		startEvent := &(*p.StartEvents())[i]
		// Start event that observes some events
		if len(startEvent.EventDefinitions()) > 0 {
			result = append(result, startEvent)
		}
	}

	for i := range *p.EventBasedGateways() {
		gateway := &(*p.EventBasedGateways())[i]
		// Event-based gateways with `instantiate` set to true
		// and no incoming sequence flows
		if gateway.Instantiate() && len(*gateway.Incomings()) == 0 {
			result = append(result, gateway)
		}
	}

	for i := range *p.ReceiveTasks() {
		task := &(*p.ReceiveTasks())[i]
		// Event-based gateways with `instantiate` set to true
		// and no incoming sequence flows
		if task.Instantiate() && len(*task.Incomings()) == 0 {
			result = append(result, task)
		}
	}

	return
}
