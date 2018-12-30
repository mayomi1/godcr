package desktop

import (
	"image"
	"image/color"
	"image/draw"
	"strconv"

	"github.com/aarzilli/nucular"
	"github.com/aarzilli/nucular/label"
	"github.com/aarzilli/nucular/rect"
	"github.com/raedahgroup/godcr/walletrpcclient"
	qrcode "github.com/skip2/go-qrcode"
)

var (
	err error

	// walletrpcclient responses
	accountBalanceResponse  []*walletrpcclient.AccountBalanceResult
	accountsResponse        []*walletrpcclient.AccountBalanceResult
	generateAddressResponse *walletrpcclient.ReceiveResult
	transactionsResponse    []*walletrpcclient.Transaction
	utxosResponse           []*walletrpcclient.UnspentOutputsResult

	// form values
	sendAmount  float64
	sendAddress string

	// form inputs
	amountInput   nucular.TextEditor
	addressInput  nucular.TextEditor
	passwordInput nucular.TextEditor

	// form selector index
	selectedAccountIndex  = 0
	selectedAccountNumber = uint32(0)

	// selected values
	selectedUTXOS []string

	// form checkbox values
	checkedUTXOS []bool

	// passphrase modal bounds
	passphraseModalBounds = rect.Rect{
		H: 200,
		W: 250,
		X: 270,
		Y: 50,
	}
)

func resetVars() {
	err = nil
	accountBalanceResponse = nil
	accountsResponse = nil
	generateAddressResponse = nil
	transactionsResponse = nil
	sendAmount = 0.0
	sendAddress = ""
	selectedAccountIndex = 0
	selectedAccountNumber = uint32(0)
	selectedUTXOS = nil
	checkedUTXOS = nil
}

func (d *Desktop) BalanceHandler(w *nucular.Window) {
	// check if already fetched. If so, do not fetch again
	if accountBalanceResponse == nil && err == nil {
		accountBalanceResponse, err = d.walletClient.Balance()
	}

	// draw page
	if page := newWindow("Balance Page", w, 0); page != nil {
		page.header("Balance")

		// content area
		if content := page.contentWindow("Balance Content"); content != nil {
			if err != nil {
				content.setErrorMessage(err.Error())
			} else {
				content.Row(20).Ratio(0.12, 0.12, 0.15, 0.15, 0.26, 0.20)
				content.Label("Account", "LC")
				content.Label("Total", "LC")
				content.Label("Spendable", "LC")
				content.Label("Locked", "LC")
				content.Label("Voting Authority", "LC")
				content.Label("Unconfirmed", "LC")

				// rows
				for _, v := range accountBalanceResponse {
					content.Label(v.AccountName, "LC")
					content.Label(strconv.Itoa(int(v.Total/100000000)), "LC")
					content.Label(strconv.Itoa(int(v.Spendable/100000000)), "LC")
					content.Label(strconv.Itoa(int(v.LockedByTickets/100000000)), "LC")
					content.Label(strconv.Itoa(int(v.VotingAuthority/100000000)), "LC")
					content.Label(strconv.Itoa(int(v.Unconfirmed/100000000)), "LC")
				}
			}
			content.end()
		}
		page.end()
	}
}

func (d *Desktop) TransactionsHandler(w *nucular.Window) {
	if transactionsResponse == nil && err == nil {
		transactionsResponse, err = d.walletClient.GetTransactions()
	}

	if page := newWindow("Transactions Page", w, 0); page != nil {
		page.header("Transactions")

		// content area
		if content := page.contentWindow("Transactions Content"); content != nil {

			if err != nil {
				content.setErrorMessage(err.Error())
			} else {
				content.Row(20).Ratio(0.18, 0.12, 0.1, 0.15, 0.15, 0.3)
				content.Label("Date", "LC")
				content.Label("Amount", "LC")
				content.Label("Fee", "LC")
				content.Label("Direction", "LC")
				content.Label("Type", "LC")
				content.Label("Hash", "LC")

				for _, tx := range transactionsResponse {
					content.Row(20).Ratio(0.18, 0.12, 0.1, 0.15, 0.15, 0.3)
					amount := amountToString(tx.Amount)
					fee := amountToString(tx.Fee)

					content.Label(tx.FormattedTime, "LC")
					content.Label(amount, "LC")
					content.Label(fee, "LC")
					content.Label(tx.Direction.String(), "LC")
					content.Label(tx.Type, "LC")
					content.Label(tx.Hash, "LC")
				}
			}
			content.end()
		}
		page.end()
	}
}

// subpage belonging to ReceiveHandler
func (d *Desktop) generateAddressHandler(w *nucular.Window) {
	if page := newWindow("Generate Address Page", w, 0); page != nil {
		page.header("Generate Address Result")

		// content area
		if content := page.contentWindow("Generate Address Result Content"); content != nil {
			content.Row(50).Dynamic(1)
			content.LabelWrap("Address: " + generateAddressResponse.Address)

			// generate qrcode
			png, err := qrcode.New(generateAddressResponse.Address, qrcode.Medium)
			if err != nil {
				content.Row(300).Dynamic(1)
				content.LabelWrap(err.Error())
			} else {
				content.Row(200).Dynamic(1)
				img := png.Image(300)
				imgRGBA := image.NewRGBA(img.Bounds())
				draw.Draw(imgRGBA, img.Bounds(), img, image.Point{}, draw.Src)
				content.Image(imgRGBA)
			}
			content.end()
		}
		page.end()
	}
}

func (d *Desktop) ReceiveHandler(w *nucular.Window) {
	// check if already fetched. If so, do not fetch again
	if accountsResponse == nil && err == nil {
		accountsResponse, err = d.walletClient.Balance()
	}

	// draw page
	if page := newWindow("ReceivePage", w, 0); page != nil {
		page.header("Receive")

		// content area
		if content := page.contentWindow("Receive Content"); content != nil {
			if err != nil {
				content.setErrorMessage(err.Error())
			} else {
				accountNames := make([]string, len(accountsResponse))
				for index, account := range accountsResponse {
					accountNames[index] = account.AccountName
				}

				content.Row(30).Ratio(0.75, 0.25)
				// draw select account combo
				selectedAccountIndex = content.ComboSimple(accountNames, selectedAccountIndex, 30)
				// draw submit button
				if content.Button(label.T("Generate"), false) {
					// get selected account by index
					accountName := accountNames[selectedAccountIndex]
					for _, account := range accountsResponse {
						if account.AccountName == accountName {
							selectedAccountNumber = account.AccountNumber
							break
						}
					}

					// get address
					if generateAddressResponse == nil && err == nil {
						generateAddressResponse, err = d.walletClient.Receive(selectedAccountNumber)
						if err != nil {
							content.setErrorMessage(err.Error())
						} else {
							d.gotoSubpage("generateaddress")
						}
					}
				}
			}
			content.end()
		}
		page.end()
	}
}

func (d *Desktop) selectUTXOSHandler(w *nucular.Window) {
	if utxosResponse == nil && err == nil {
		utxosResponse, err = d.walletClient.UnspentOutputs(selectedAccountNumber, 0)
	}

	// draw page
	if page := newWindow("Select UTXOS", w, 0); page != nil {
		page.header("Select UTXOS for custom transaction")

		if content := page.contentWindow("Select UTXOS Content"); content != nil {
			if err != nil {
				content.setErrorMessage(err.Error())
			} else {
				content.Row(20).Dynamic(1)
				content.LabelColored("Select UTXOS to create custom transaction", "LC", color.RGBA{106, 106, 106, 255})

				content.Row(230).Dynamic(1)
				if txGroup := content.GroupBegin("UTXOS", 0); txGroup != nil {
					txGroup.Row(20).Ratio(0.05, 0.7, 0.25)
					txGroup.Label("", "LC")
					txGroup.Label("Transaction Hash", "LC")
					txGroup.Label("Amount", "LC")
					//txGroup.Label("Time", "LC")

					if checkedUTXOS == nil {
						checkedUTXOS = make([]bool, len(utxosResponse))
					}

					if selectedUTXOS == nil {
						selectedUTXOS = make([]string, len(utxosResponse))
					}

					for i, v := range utxosResponse {
						if txGroup.CheckboxText("", &checkedUTXOS[i]) {
							if checkedUTXOS[i] {
								selectedUTXOS[i] = v.TransactionHash
							} else {
								selectedUTXOS[i] = ""
							}
						}
						txGroup.Label(v.TransactionHash, "LC")
						txGroup.Label(v.AmountString, "LC")
						//txGroup.Label("time", "LC")
					}
					txGroup.GroupEnd()
				}

				content.Row(80).Dynamic(1)
				if submitButtonGroup := content.GroupBegin("SubmitButtonGroup", nucular.WindowNoHScrollbar); submitButtonGroup != nil {
					submitButtonGroup.Row(50).Static(150)
					if submitButtonGroup.Button(label.T("Next"), false) {
						// get passphrase
						walletPassphrase := ""

						getPassphraseHandler := func(w *nucular.Window) {
							w.Row(25).Dynamic(1)
							w.Label("Wallet Passphrase", "LC")

							w.Row(30).Dynamic(1)
							passwordInput.Edit(w)

							w.Row(25).Dynamic(1)
							if w.Button(label.T("Submit"), false) {
								walletPassphrase = string(passwordInput.Buffer)
							}
						}
						submitButtonGroup.Master().PopupOpen("Wallet Passphrase", nucular.WindowTitle, passphraseModalBounds, true, getPassphraseHandler)
					}
					submitButtonGroup.GroupEnd()
				}

			}
			content.end()
		}
		page.end()
	}
}

func (d *Desktop) SendHandler(w *nucular.Window) {
	if accountsResponse == nil && err == nil {
		accountsResponse, err = d.walletClient.Balance()
	}

	// draw page
	if page := newWindow("Send Page", w, 0); page != nil {
		page.header("Send")

		// content area
		if content := page.contentWindow("Send Content"); content != nil {
			if err != nil {
				content.setErrorMessage(err.Error())
			} else {
				accounts := make([]string, len(accountsResponse))
				for index, account := range accountsResponse {
					accounts[index] = account.AccountName
				}

				content.Row(15).Dynamic(2)
				content.Label("Account:", "LC")
				content.Label("Amount:", "LC")

				content.Row(25).Dynamic(2)
				selectedAccountNumber = 0
				// account select input
				selectedAccountIndex = content.ComboSimple(accounts, selectedAccountIndex, 25)

				// amount text input
				amountInput.Edit(content.Window)

				content.Row(25).Dynamic(2)
				content.Label("Destination Address:", "LC")

				content.Row(25).Dynamic(2)
				// address text input
				addressInput.Edit(content.Window)

				content.Row(35).Static(300)
				if content.Button(label.T("Next"), false) {
					// TODO validation

					// get account number from selected index
					accountName := accounts[selectedAccountIndex]
					for _, account := range accountsResponse {
						if account.AccountName == accountName {
							selectedAccountNumber = account.AccountNumber
							break
						}
					}
					d.gotoSubpage("selectutxos")
				}
			}
			content.end()
		}
		page.end()
	}
}
