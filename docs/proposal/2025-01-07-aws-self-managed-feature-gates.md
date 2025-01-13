---
title: Feature gates for AWSCluster and AWSMachine
authors:
  - "@mzazrivec"
reviewers:
  - "@serngawy"
  - "@nrb"
creation-date: 2025-01-07
last-updated: 2025-01-13
status: provisional
---

# Feature gates for AWS self managed clusters

## Table of Contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Glossary](#glossary)
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals/Future Work](#non-goalsfuture-work)
- [Proposal](#proposal)
  - [User Stories](#user-stories)
    - [Functional Requirements](#functional-requirements)
- [Alternatives](#alternatives)
- [Upgrade Strategy](#upgrade-strategy)
- [Implementation History](#implementation-history)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Glossary

Refer to the [Cluster API Book Glossary](https://cluster-api.sigs.k8s.io/reference/glossary.html).

## Summary

Feature gates are a mechanism that CAPA uses to enable or disable particular features components of CAPA. This document proposes to implement two new feature gates, one for AWSMachine controller, another one for AWSCluster controller.

## Motivation

Motivation for the two new feature gates is the option to turn on and off CAPA's ability to reconcile AWSMachine and AWSCluster resources. Currently, controllers for these two resource types are always on.

The possibility to turn the respective controllers on and off becomes important in multi-tenant CAPA installations (more than one CAPA installed in a single kubernetes cluster). The two new feature gates will help avoid interference and conflicts between the controllers during resource reconciliations.

### Goals

1. Implement a feature gate for AWSMachine controller
2. Implement a feature gate for AWSCluster controller

### Non-Goals/Future Work

- change any of the existing feature gates or their semantics

## Proposal

- Introduce AWSMachine feature gate which would turn on and off the AWSMachineReconciler. The feature gate would be on by default.
- Introduce AWSCluster feature gate which would turn on and off the AWSClusterReconciler. The feature gate would be on by default.

### User Stories

As a CAPA user and a cluster admin, I want to be able to install two (or more) CAPA instances on my cluster. The first instance would have the feature gates for AWS self managed cluster enabled, the other instances would have those feature gates disabled.

#### Functional Requirements

1. Ability to enable / disable feature gates for AWSCluster and AWSMachine controllers.
2. Both feature gates would be on by default.

## Alternatives

The alternative here would be simply not to implement the two new feature gates and rely just on namespace separation.

## Upgrade Strategy

CAPA upgrades should not be affected. Existing deployments will not notice anything different, because the two new feature gates will be on by default (current behavior).

## Implementation History

- [ ] 2025-01-20: Proposed idea in an issue or [community meeting]

<!-- Links -->
[community meeting]: https://docs.google.com/document/d/1ushaVqAKYnZ2VN_aa3GyKlS4kEd6bSug13xaXOakAQI/edit#heading=h.pxsq37pzkbdq
