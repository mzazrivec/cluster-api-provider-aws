/*
Copyright 2019 The Kubernetes Authors.

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

// Package mock_s3iface provides a mock implementation of the S3API interface
// Run go generate to regenerate this mock.
//
//go:generate ../../../../../hack/tools/bin/mockgen -destination s3api_mock.go -package mock_s3iface sigs.k8s.io/cluster-api-provider-aws/v2/pkg/cloud/services/s3 S3API
//go:generate /usr/bin/env bash -c "cat ../../../../../hack/boilerplate/boilerplate.generatego.txt s3api_mock.go > _s3api_mock.go && mv _s3api_mock.go s3api_mock.go"
package mock_s3iface //nolint:stylecheck
