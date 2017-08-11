package acceptance_test

import (
	"bytes"
	"os"
	"os/exec"
	"time"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	"github.com/cloudfoundry/bosh-bootloader/acceptance-tests/actors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = FDescribe("up test", func() {
	var (
		bbl     actors.BBL
		bosh    actors.BOSH
		boshcli actors.BOSHCLI
		state   acceptance.State
	)

	// AfterEach(func() {
	// 	bbl.Destroy()
	// })

	It("checks the credentials", func() {
		var err error
		config, err := acceptance.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		bbl = actors.NewBBL(config.StateFileDir, pathToBBL, config, "azure-env")
		bosh = actors.NewBOSH()
		boshcli = actors.NewBOSHCLI()
		state = acceptance.NewState(config.StateFileDir)

		args := []string{
			"--state-dir", config.StateFileDir,
			"--debug",
			"up",
		}

		args = append(args, []string{
			"--iaas", "azure",
			"--azure-subscription-id", config.AzureSubscriptionID,
			"--azure-tenant-id", config.AzureTenantID,
			"--azure-client-id", config.AzureClientID,
			"--azure-client-secret", config.AzureClientSecret,
		}...)

		cmd := exec.Command(pathToBBL, args...)
		stdout := bytes.NewBuffer([]byte{})
		session, err := gexec.Start(cmd, stdout, os.Stderr)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 40*time.Minute).Should(gexec.Exit(0))

		Expect(string(stdout.Bytes())).To(ContainSubstring("step: verifying credentials"))
	})
})
