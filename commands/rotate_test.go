package commands_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rotate", func() {
	var (
		stateStore       *fakes.StateStore
		keyPairManager   *fakes.KeyPairManager
		terraformManager *fakes.TerraformManager
		boshManager      *fakes.BOSHManager
		stateValidator   *fakes.StateValidator

		command commands.Rotate

		incomingState storage.State
	)

	BeforeEach(func() {
		stateStore = &fakes.StateStore{}
		keyPairManager = &fakes.KeyPairManager{}
		terraformManager = &fakes.TerraformManager{}
		boshManager = &fakes.BOSHManager{}
		stateValidator = &fakes.StateValidator{}

		command = commands.NewRotate(stateStore, keyPairManager, terraformManager, boshManager, stateValidator)
	})

	Describe("CheckFastFails", func() {
		It("returns an error when state validator fails", func() {
			stateValidator.ValidateCall.Returns.Error = errors.New("state validator failed")
			err := command.CheckFastFails([]string{}, incomingState)
			Expect(err).To(MatchError("state validator failed"))
		})
	})

	Describe("Execute", func() {
		BeforeEach(func() {
			incomingState = storage.State{
				KeyPair: storage.KeyPair{
					Name:       "some-name",
					PrivateKey: "some-private-key",
					PublicKey:  "some-public-key",
				},
			}

			keyPairManager.RotateCall.Returns.KeyPair = storage.KeyPair{
				Name:       "some-new-name",
				PrivateKey: "some-new-private-key",
				PublicKey:  "some-new-public-key",
			}

			boshManager.CreateDirectorCall.Returns.State = storage.State{
				KeyPair: storage.KeyPair{
					Name:       "some-new-name",
					PrivateKey: "some-new-private-key",
					PublicKey:  "some-new-public-key",
				},
				BOSH: storage.BOSH{
					DirectorName: "some-director-name",
				},
			}
		})

		It("rotates the keys", func() {
			err := command.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(keyPairManager.RotateCall.CallCount).To(Equal(1))
			Expect(keyPairManager.RotateCall.Receives.State).To(Equal(incomingState))
			Expect(stateStore.SetCall.CallCount).To(BeNumerically(">=", 1))
			Expect(stateStore.SetCall.Receives[0].State).To(Equal(storage.State{
				KeyPair: storage.KeyPair{
					Name:       "some-new-name",
					PrivateKey: "some-new-private-key",
					PublicKey:  "some-new-public-key",
				},
			}))
			Expect(terraformManager.GetOutputsCall.CallCount).To(Equal(1))
			Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(storage.State{
				KeyPair: storage.KeyPair{
					Name:       "some-new-name",
					PrivateKey: "some-new-private-key",
					PublicKey:  "some-new-public-key",
				},
			}))
		})

		It("redeploys bosh", func() {
			err := command.Execute([]string{}, incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(1))
			Expect(boshManager.CreateDirectorCall.Receives.State).To(Equal(storage.State{
				KeyPair: storage.KeyPair{
					Name:       "some-new-name",
					PrivateKey: "some-new-private-key",
					PublicKey:  "some-new-public-key",
				},
			}))

			Expect(stateStore.SetCall.CallCount).To(BeNumerically(">=", 2))
			Expect(stateStore.SetCall.Receives[1].State).To(Equal(storage.State{
				KeyPair: storage.KeyPair{
					Name:       "some-new-name",
					PrivateKey: "some-new-private-key",
					PublicKey:  "some-new-public-key",
				},
				BOSH: storage.BOSH{
					DirectorName: "some-director-name",
				},
			}))
		})

		Context("when no director exists", func() {
			BeforeEach(func() {
				incomingState.NoDirector = true
			})

			It("does not deploy bosh", func() {
				err := command.Execute([]string{}, incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshManager.CreateDirectorCall.CallCount).To(Equal(0))
				Expect(stateStore.SetCall.CallCount).To(Equal(1))
			})
		})

		Context("failure cases", func() {
			It("returns an error when key pair manager rotate fails", func() {
				keyPairManager.RotateCall.Returns.Error = errors.New("failed to rotate")
				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to rotate"))
			})

			It("returns an error when stateStore set fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{errors.New("failed to set")}}
				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to set"))
			})

			It("returns an error when terraformManager getOutputs fails", func() {
				terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to get outputs")
				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to get outputs"))
			})

			It("returns an error when boshManager create fails", func() {
				boshManager.CreateDirectorCall.Returns.Error = errors.New("failed to create")
				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to create"))
			})

			It("returns an error when stateStore set fails", func() {
				stateStore.SetCall.Returns = []fakes.SetCallReturn{{}, {errors.New("failed to set")}}
				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("failed to set"))
			})
		})
	})
})
