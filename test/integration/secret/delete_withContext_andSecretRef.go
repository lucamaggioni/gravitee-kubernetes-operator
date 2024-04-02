// Copyright (C) 2015 The Gravitee team (http://gravitee.io)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package secret_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gravitee-io/gravitee-kubernetes-operator/pkg/keys"
	"github.com/gravitee-io/gravitee-kubernetes-operator/test/internal/integration/assert"
	"github.com/gravitee-io/gravitee-kubernetes-operator/test/internal/integration/constants"
	"github.com/gravitee-io/gravitee-kubernetes-operator/test/internal/integration/fixture"
	"github.com/gravitee-io/gravitee-kubernetes-operator/test/internal/integration/labels"
	"github.com/gravitee-io/gravitee-kubernetes-operator/test/internal/integration/manager"
	"k8s.io/apimachinery/pkg/api/errors"
)

var _ = Describe("Delete", labels.WithContext, func() {
	timeout := constants.EventualTimeout

	interval := constants.Interval

	ctx := context.Background()

	It("should delete only with no more context references", func() {
		fixtures := fixture.Builder().
			AddSecret(constants.ContextSecretFile).
			WithContext(constants.ContextWithSecretFile).
			Build()

		secret := fixtures.Secrets[0]

		secret.Name += fixtures.GetGeneratedSuffix()
		fixtures.Context.Spec.Auth.SecretRef.Name = secret.Name

		fixtures.Apply()

		By("expecting finalizer to have been added to secret")

		Eventually(func() error {
			sec, err := manager.GetLatest(secret)
			if err != nil {
				return err
			}
			return assert.AssertFinalizer(sec, keys.ManagementContextSecretFinalizer)
		}, timeout, interval).Should(Succeed())

		By("deleting secret")

		Expect(manager.Client().Delete(ctx, secret)).To(Succeed())

		By("expecting to still find secret")

		checkUntil := constants.ConsistentTimeout
		Consistently(func() error {
			_, kErr := manager.GetLatest(secret)
			return kErr
		}, checkUntil, interval).Should(Succeed())

		By("deleting the management context")

		Expect(manager.Client().Delete(ctx, fixtures.Context)).To(Succeed())

		By("expecting secret to have been deleted")

		Eventually(func() error {
			_, kErr := manager.GetLatest(secret)
			if errors.IsNotFound(kErr) {
				return nil
			}
			return assert.Equals("error", "[NOT FOUND]", kErr)
		}, timeout, interval).Should(Succeed())
	})
})
