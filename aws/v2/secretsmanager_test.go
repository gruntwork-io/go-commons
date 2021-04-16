package aws

import (
	"testing"

	goaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test GetSecretsManagerSecret by creating the secret in terratest and using the function to read the value to make
// sure it reads the correct secret value.
func TestGetSecretString(t *testing.T) {
	t.Parallel()

	name := "refarch-deployer-test-secret-" + random.UniqueId()
	secretVal := random.UniqueId()
	region := aws.GetRandomStableRegion(t, nil, nil)
	opts := NewOptions(region)

	secretARN := aws.CreateSecretStringWithDefaultKey(t, region, "Test Secret for refarch-deployer", name, secretVal)
	defer aws.DeleteSecret(t, region, secretARN, true)

	actualSecret, err := GetSecretsManagerSecretString(opts, secretARN)
	require.NoError(t, err)
	assert.Equal(t, secretVal, actualSecret)
}

// Test SecretsManagerEntryExists returns false when calling it on a secrets manager entry that doesn't exist.
func TestSecretsManagerEntryExistsFalse(t *testing.T) {
	t.Parallel()

	opts := NewOptions("us-east-1")
	exists, err := SecretsManagerEntryExists(opts, "secret-that-doesnt-exist")
	require.NoError(t, err)
	assert.False(t, exists)
}

// Test SecretsManagerEntryExists returns true when calling it on a secrets manager entry that exists.
func TestSecretsManagerEntryExistsTrueWithARN(t *testing.T) {
	t.Parallel()

	name := "refarch-deployer-test-secret-" + random.UniqueId()
	secretVal := random.UniqueId()
	region := aws.GetRandomStableRegion(t, nil, nil)
	opts := NewOptions(region)

	secretARN := aws.CreateSecretStringWithDefaultKey(t, region, "Test Secret for refarch-deployer", name, secretVal)
	defer aws.DeleteSecret(t, region, secretARN, true)

	exists, err := SecretsManagerEntryExists(opts, secretARN)
	require.NoError(t, err)
	assert.True(t, exists)
}

// Test SecretsManagerEntryExists returns true when calling it on a secrets manager entry that exists with name.
func TestSecretsManagerEntryExistsTrueWithName(t *testing.T) {
	t.Parallel()

	name := "refarch-deployer-test-secret-" + random.UniqueId()
	secretVal := random.UniqueId()
	region := aws.GetRandomStableRegion(t, nil, nil)
	opts := NewOptions(region)

	secretARN := aws.CreateSecretStringWithDefaultKey(t, region, "Test Secret for refarch-deployer", name, secretVal)
	defer aws.DeleteSecret(t, region, secretARN, true)

	secret, err := GetSecretsManagerMetadata(opts, secretARN)
	require.NoError(t, err)

	exists, err := SecretsManagerEntryExists(opts, goaws.ToString(secret.Name))
	require.NoError(t, err)
	assert.True(t, exists)
}
