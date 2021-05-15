<img src="https://github.com/bpxe/bpxe/blob/master/logo.svg" width="100%" height="150">

# BPXE: Business Process eXecution Engine

BPMN 2.0 based business process execution engine implemented in Go. BPMN stands
for Business Process Model and Notation. BPMN's goal is to help stakeholders to
have a shared understanding of processes.

BPXE focuses on the execution aspect of such notation, effectively allowing the
processes described in BPMN to function as if they were programs. BPXE is not
the only such engine, as there are many commercially or community supported
ones.

Having visualized processes that are also determinisitcally executable is key to
maintaining a coherent understanding of what the process is supposed to do across
teams and specialties.

## Processes as Source of Truth

Equally important aspect of BPXE is that it makes processes and their executions
a durable source of truth. This means that process instances can query previous
executions of any processes to make further decisions.

As an example, consider a Purchasing process, which chooses a special path if it
queries previous execution and finds out that when given the same or similar shopping
cart, majority of those process executions were abandoned (i.e. customer did not complete
a purchase). This kind of logic can be easily integrated into a process and updated as needed,
giving a much better level of insight and control at a much lower modification cost.

## Goals

* Reasonably good performance
* Small footprint
* Multiplatform (servers, browsers, microcontrollers)
* Multitenancy capabilities
* Distributed process execution
* Semantic correctness
* Failure resistance

## Usage

At this time, BPXE does not have an executable server of its own and can be only used as a Go library.

## Licensing & Contributions

BPXE is Open Source software in the making. Its source code is currently
available under the terms of [Business Source License 1.1](LICENSE) with an
Additional Use Grant for non-commercial purposes. Moreover, according to the
terms of this license, every release of BPXE will eventually be re-licensed
under the terms of [Apache 2.0 license](licenses/LICENSE-Apache-2.0), on its
fourth anniversary.

We also take [contributions](CONTRIBUTING.md) under the terms of [New BSD
license](licenses/LICENSE-BSD-3-Clause) or in public domain.

