package ssh

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/aws"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	maxRetries         = 30
	timeBetweenRetries = 5 * time.Second
)

func TestSSHRunCommand(t *testing.T) {
	t.Parallel()

	//os.Setenv("SKIP_setup", "true")
	//os.Setenv("SKIP_deploy", "true")
	//os.Setenv("SKIP_validate", "true")
	//os.Setenv("SKIP_teardown", "true")

	workingDir := filepath.Join(".", "stages", t.Name())

	defer test_structure.RunTestStage(t, "teardown", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)
		terraform.Destroy(t, terraformOptions)

		keyPair := test_structure.LoadEc2KeyPair(t, workingDir)
		aws.DeleteEC2KeyPair(t, keyPair)
	})

	test_structure.RunTestStage(t, "setup", func() {
		terraformOptions, keyPair := configureTerraformOptions(t)

		test_structure.SaveTerraformOptions(t, workingDir, terraformOptions)
		test_structure.SaveEc2KeyPair(t, workingDir, keyPair)
	})

	test_structure.RunTestStage(t, "deploy", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)
		terraform.InitAndApply(t, terraformOptions)
	})

	// Make sure we can SSH to the public Instance directly from the public Internet and the private Instance by using
	// the public Instance as a jump host
	test_structure.RunTestStage(t, "validate", func() {
		terraformOptions := test_structure.LoadTerraformOptions(t, workingDir)
		keyPair := test_structure.LoadEc2KeyPair(t, workingDir)

		// Run `terraform output` to get the value of an output variable
		publicInstanceIP := terraform.Output(t, terraformOptions, "public_instance_ip")
		publicInstancePrivateIP := terraform.Output(t, terraformOptions, "public_instance_private_ip")
		privateInstanceIP := terraform.Output(t, terraformOptions, "private_instance_ip")
		privateRestrictedInstanceIP := terraform.Output(t, terraformOptions, "private_restricted_instance_ip")

		// Set up the Host struct to connect to each instance, using the required jump host for accessing the nested
		// instances.
		publicHost := Host{
			Hostname:        publicInstanceIP,
			PrivateKey:      keyPair.KeyPair.PrivateKey,
			SSHUserName:     "ubuntu",
			HostKeyCallback: NoOpHostKeyCallback,
		}
		privateHost := Host{
			Hostname:        privateInstanceIP,
			PrivateKey:      keyPair.KeyPair.PrivateKey,
			SSHUserName:     "ubuntu",
			JumpHost:        &publicHost,
			HostKeyCallback: NoOpHostKeyCallback,
		}
		privateRestrictedHost := Host{
			Hostname:        privateRestrictedInstanceIP,
			PrivateKey:      keyPair.KeyPair.PrivateKey,
			SSHUserName:     "ubuntu",
			JumpHost:        &privateHost,
			HostKeyCallback: NoOpHostKeyCallback,
		}

		testCases := []struct {
			name        string
			description string
			host        Host
			checkIP     string
		}{
			{
				"publichost",
				fmt.Sprintf("SSH to public host with IP %s", publicInstanceIP),
				publicHost,
				publicInstancePrivateIP,
			},
			{
				"privatehost",
				fmt.Sprintf("SSH to private host with IP %s using public host %s", privateInstanceIP, publicInstanceIP),
				privateHost,
				privateInstanceIP,
			},
			{
				"privateRestrictedHost",
				fmt.Sprintf("SSH to private restricted host with IP %s using private host %s with public host %s", privateRestrictedInstanceIP, privateInstanceIP, publicInstanceIP),
				privateRestrictedHost,
				privateRestrictedInstanceIP,
			},
		}

		// Insulate each test case in a synchronous wrapper test so that all tests run to completion before we continue.
		t.Run("group", func(t *testing.T) {
			for _, testCase := range testCases {
				// Capture the range variable so that it doesn't change when we spawn each test and the goroutine swaps.
				testCase := testCase
				t.Run(testCase.name, func(t *testing.T) {
					t.Parallel()
					testSSHToHost(t, testCase.host, testCase.description, testCase.checkIP)
				})
			}

			t.Run("publichostWithAgent", func(t *testing.T) {
				t.Parallel()

				sshAgent, err := SSHAgentWithPrivateKey(keyPair.KeyPair.PrivateKey)
				defer sshAgent.Stop()
				require.NoError(t, err)
				publicHostCopy := publicHost
				publicHostCopy.OverrideSSHAgent = sshAgent
				testSSHToHost(t, publicHostCopy, "SSH to public host with ssh-agent", publicInstancePrivateIP)
			})
		})
	})
}

func configureTerraformOptions(t *testing.T) (*terraform.Options, *aws.Ec2Keypair) {
	uniqueID := random.UniqueId()

	instanceName := fmt.Sprintf("terratest-ssh-example-%s", uniqueID)

	awsRegion := aws.GetRandomStableRegion(t, nil, nil)

	keyPairName := fmt.Sprintf("terratest-ssh-example-%s", uniqueID)
	keyPair := aws.CreateAndImportEC2KeyPair(t, awsRegion, keyPairName)

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "./test-fixture/module",

		Vars: map[string]interface{}{
			"aws_region":    awsRegion,
			"instance_name": instanceName,
			"key_pair_name": keyPairName,
		},
	})

	return terraformOptions, keyPair
}

func testSSHToHost(t *testing.T, host Host, description string, checkIP string) {
	cmdOut := retry.DoWithRetry(t, description, maxRetries, timeBetweenRetries, func() (string, error) {
		return RunCommandAndGetStdout(host, "ip a")
	})
	assert.Contains(t, cmdOut, checkIP)
}
