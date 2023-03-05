package chain_sdk

import (	
	"fmt"
	"github.com/stellar/go/txnbuild"
	"strings"

	"encoding/json"
	"errors"
	"github.com/stellar/go/keypair"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"net/http"
)

const (
	PI_NETWORK_PLATFORM_HOST   = "https://api.minepi.com"
	PI_NETWORK_CHAIN_HOST      = "https://api.mainnet.minepi.com/"
	PI_NETWORK_TEST_CHAIN_HOST = "https://api.testnet.minepi.com"
	PI_NETWORK_PASSPHRASE      = "Pi Network"
	PI_NETWORK_TEST_PASSPHRASE = "Pi Testnet"
	PI_API_KEY                 = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
)

type PiNetworkSDK struct {
	DevMode int
	BaseUrl string
	//OpenPayment       map[string]interface{}
	Headers           map[string]string
	HorizonURL        string
	NetworkPassphrase string
	secret            string
	uid               string
	kp                keypair.KP
	accountDetail     hProtocol.Account
	feeStats          hProtocol.FeeStats
	client            horizonclient.Client
}

func checkPrivateSeedValid(seed string) bool {
	upperSeed := strings.ToUpper(seed)
	if len(seed) != 56 || !strings.HasPrefix(upperSeed, "S") {
		return false
	}
	return true
}
func NewPiNetworkSDK() (*PiNetworkSDK, error) {
	pisdk := PiNetworkSDK{DevMode: 1, BaseUrl: PI_NETWORK_PLATFORM_HOST, NetworkPassphrase: PI_NETWORK_PASSPHRASE}
	pisdk.Headers = make(map[string]string, 0)
	pisdk.Headers["Authorization"] = " Key " + PI_API_KEY

	horizonURL := PI_NETWORK_CHAIN_HOST
	if pisdk.DevMode == 1 {
		horizonURL = PI_NETWORK_TEST_CHAIN_HOST
		pisdk.NetworkPassphrase = PI_NETWORK_TEST_PASSPHRASE
	}
	pisdk.HorizonURL = horizonURL
	pisdk.client = horizonclient.Client{
		HorizonURL: horizonURL,
		HTTP:       http.DefaultClient,
	}
	pisdk.client.SetHorizonTimeout(horizonclient.HorizonTimeout)

	feeStats, err := pisdk.client.FeeStats()
	if err != nil {
		return nil, err
	}
	pisdk.feeStats = feeStats
	return &pisdk, nil
}


func (t *PiNetworkSDK) LoadAccountDetails(secret string) error {
	if !checkPrivateSeedValid(secret) {
		return errors.New("Invalid Secret")
	}

	kp, err := keypair.Parse(secret)
	if err != nil {
		return err
	}

	ar := horizonclient.AccountRequest{AccountID: kp.Address()}
	accountDetail, err := t.client.AccountDetail(ar)
	if err != nil {
		return err
	}
	t.kp = kp
	t.accountDetail = accountDetail
	return nil
}

func (t *PiNetworkSDK) GetBalance() string {
	balances := t.accountDetail.Balances
	for i := range balances {
		if balances[i].Asset.Type == "native" {
			return balances[i].Balance
		}
	}
	return "0"
}

func (t *PiNetworkSDK) GetIncompletePayments() ([]interface{}, error) {
	url := t.BaseUrl + "/v2/payments/incomplete_server_payments"
	//data := make(map[string]interface{}, 0)
	bytes, err := http2.GetWithHeaders(url, t.Headers, 30)
	if err != nil {
		return nil, err
	}
	var resultmap map[string]interface{}
	err = json.Unmarshal(bytes, &resultmap)
	if err != nil {
		return nil, err
	}
	return resultmap["incomplete_server_payments"].([]interface{}), nil
}

func (t *PiNetworkSDK) GetPayment(paymentID string) (map[string]interface{}, error) {
	url := t.BaseUrl + "/v2/payments/paymentID"
	bytes, err := http2.PostWithHeaders(url, nil, t.Headers, 30)
	if err != nil {
		return nil, err
	}
	var resultmap map[string]interface{}
	err = json.Unmarshal(bytes, &resultmap)
	if err != nil {
		return nil, err
	}
	return resultmap, nil
}

func (t *PiNetworkSDK) CreatePayment(uid string, amount string, title string, reference string) (map[string]interface{}, error) {
	url := t.BaseUrl + "/v2/payments"
	postmap := make(map[string]interface{})
	postmap["uid"] = uid
	postmap["amount"] = amount
	//postmap["direction"] = "user_to_app"
	postmap["memo"] = title
	postmap["metadata"] = map[string]interface{}{"reference": reference}
	payment := map[string]interface{}{"payment": postmap}
	//headers["Content-Type"] = "application/json"
	bytes, err := http2.PostWithHeaders(url, payment, t.Headers, 30)
	if err != nil {
		return nil, err
	}
	var resultMap map[string]interface{}
	err = json.Unmarshal(bytes, &resultMap)
	if err != nil {
		return nil, err
	}
	if resultMap["error"] != "" && resultMap["error"] != nil {
		return nil, errors.New(resultMap["error"].(string))
	}
	return resultMap, nil
}

func (t *PiNetworkSDK) CompletePayment(paymentIdentifier string, txid string) (map[string]interface{}, error) {
	url := t.BaseUrl + "/v2/payments/" + paymentIdentifier + "/complete"
	postMap := make(map[string]interface{})
	postMap["txid"] = txid
	bytes, err := http2.PostWithHeaders(url, postMap, t.Headers, 30)
	if err != nil {
		return nil, err
	}
	var resultMap map[string]interface{}
	err = json.Unmarshal(bytes, &resultMap)
	if err != nil {
		return nil, err
	}
	if resultMap["error"] != "" && resultMap["error"] != nil {
		return nil, errors.New(resultMap["error"].(string))
	}
	return resultMap, nil
}

func (t *PiNetworkSDK) ApprovePayment(paymentIdentifier string) (map[string]interface{}, error) {
	url := t.BaseUrl + "/v2/payments/" + paymentIdentifier + "/approve"
	//data:=make(map[string]interface{},0)
	bytes, err := http2.PostWithHeaders(url, nil, t.Headers, 30)
	if err != nil {
		return nil, err
	}
	var resultMap map[string]interface{}
	err = json.Unmarshal(bytes, &resultMap)
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

func (t *PiNetworkSDK) CancelPayment(paymentIdentifier string) (map[string]interface{}, error) {
	url := t.BaseUrl + "/v2/payments/" + paymentIdentifier + "/cancel"
	//data:=make(map[string]interface{},0)
	bytes, err := http2.PostWithHeaders(url, nil, t.Headers, 30)
	if err != nil {
		return nil, err
	}
	var resultMap map[string]interface{}
	err = json.Unmarshal(bytes, &resultMap)
	if err != nil {
		return nil, err
	}

	return resultMap, nil
}

func (t *PiNetworkSDK) SubmitPayment(payment map[string]interface{}) (string, error) {
	//from_address := payment["from_address"].(string)
	amountValue := payment["amount"].(interface{})
	amount := fmt.Sprintf("%f", amountValue)
	transaction_data := map[string]interface{}{
		"to_address": payment["to_address"].(string),
		"amount":     amount,
		"identifier": payment["identifier"].(string),
	}
	transaction, err := t.BuildTransaction(transaction_data)
	if err != nil {
		return "", err
	}
	txid, err := t.SubmitTransaction(transaction)
	if err != nil {
		return "", err
	}
	return txid, nil
}

func (t *PiNetworkSDK) BuildTransaction(transaction_data map[string]interface{}) (*txnbuild.Transaction, error) {
	amount := transaction_data["amount"].(string)
	toAddress := transaction_data["to_address"].(string)
	identifier := transaction_data["identifier"].(string)
	//from_address := transaction_data["from_address"].(string)
	op := txnbuild.Payment{
		Destination: toAddress,
		Amount:      amount,
		Asset:       txnbuild.NativeAsset{},
	}
	memoText := txnbuild.MemoText(identifier)
	// Construct the transaction that holds the operations to execute on the network
	tx, err := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount:        &t.accountDetail,
			IncrementSequenceNum: true,
			Operations:           []txnbuild.Operation{&op},
			BaseFee:              t.feeStats.MaxFee.Min,
			Memo:                 memoText,
			Preconditions:        txnbuild.Preconditions{TimeBounds: txnbuild.NewTimeout(300)},
		},
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
func (t *PiNetworkSDK) SubmitTransaction(tx *txnbuild.Transaction) (string, error) {
	var err error
	tx, err = tx.Sign(t.NetworkPassphrase, t.kp.(*keypair.Full))
	if err != nil {
		return "", err
	}

	// Get the base 64 encoded transaction envelope
	txe, err := tx.Base64()
	if err != nil {
		return "", err
	}

	resp, err := t.client.SubmitTransactionXDR(txe)
	if err != nil {
		return "", err
	}
	if !resp.Successful {
		return "", errors.New("SubmitTransactionXDR fail")
	}

	return resp.ID, nil
}

func (t *PiNetworkSDK) CreateNewAccount(fromSeed string) (string, string, error) {
	kp, _ := keypair.Parse(fromSeed)
	ar := horizonclient.AccountRequest{AccountID: kp.Address()}
	sourceAccount, err := t.client.AccountDetail(ar)
	if err != nil {
		return "", "", err
	}

	newKp, _ := keypair.Random()
	newAddress := newKp.Address()
	newSeed := newKp.Seed()

	op := txnbuild.CreateAccount{
		Destination:   newAddress,
		Amount:        "20", //20在测试网是最小值
		SourceAccount: kp.Address(),
	}
	tx, err := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount:        &sourceAccount,
			IncrementSequenceNum: true,
			Operations:           []txnbuild.Operation{&op},
			BaseFee:              t.feeStats.MaxFee.Min,
			Preconditions:        txnbuild.Preconditions{TimeBounds: txnbuild.NewInfiniteTimeout()}, // Use a real timeout in production!
		},
	)

	tx, _ = tx.Sign(t.NetworkPassphrase, kp.(*keypair.Full))
	txe, _ := tx.Base64()

	_, err = t.client.SubmitTransactionXDR(txe)
	if err != nil {
		return "", "", err
	}
	newAr := horizonclient.AccountRequest{AccountID: newAddress}
	newAccount, err := t.client.AccountDetail(newAr)
	if err != nil {
		return "", "", err
	}
	log.Info("", newAccount.AccountID)
	return newAddress, newSeed, nil
}

func (t *PiNetworkSDK) Transfer(fromSecret string, toAddress string, amount string, reference string) (string, error) {
	err := t.LoadAccountDetails(fromSecret)
	if err != nil {
		return "", err
	}
	payment := map[string]interface{}{
		"to_address": toAddress,
		"amount":     10.0,
		"identifier": "identifier",
	}
	txid, err := t.SubmitPayment(payment)
	return txid, err
}
func (t *PiNetworkSDK) GetCompletedRecharge(paymentID string) (map[string]interface{}, error) {
	return t.GetPayment(paymentID)
}

func (t *PiNetworkSDK) RechargeBalance(fromSecret string, toAddress string, amount string, reference string) (string, error) {
	err := t.LoadAccountDetails(fromSecret)
	if err != nil {
		return "", err
	}
	transaction_data := map[string]interface{}{
		"to_address": toAddress,
		"amount":     amount,
		"identifier": reference,
	}
	transaction, err := t.BuildTransaction(transaction_data)
	if err != nil {
		return "", err
	}
	txid, err := t.SubmitTransaction(transaction)
	if err != nil {
		return "", err
	}
	return txid, nil
}

func (t *PiNetworkSDK) GetInCompletedWithdraw() ([]interface{}, error) {
	return t.GetIncompletePayments()
}

func (t *PiNetworkSDK) CommitWithdraw(fromSecret string, data map[string]interface{}) (map[string]interface{}, error) {
	err := t.LoadAccountDetails(fromSecret)
	if err != nil {
		return nil, err
	}
	if data["transaction"] == nil {
		paymentIdentifier := data["identifier"].(string)
		txid, err := t.SubmitPayment(data)
		if err != nil {
			return nil, err
		}
		return t.CompletePayment(paymentIdentifier, txid)
	} else {
		transaction := data["transaction"].(map[string]interface{})
		return t.CompletePayment(data["identifier"].(string), transaction["txid"].(string))
	}
}

// uid string, amount string, title string, reference string
func (t *PiNetworkSDK) CreateWithdraw(fromSecret string, toAddressOrUid string, amount string, title string, reference string) (map[string]interface{}, error) {
	err := t.LoadAccountDetails(fromSecret)
	if err != nil {
		return nil, err
	}
	return t.CreatePayment(toAddressOrUid, amount, title, reference)
}

func (t *PiNetworkSDK) CancelWithdraw(data map[string]interface{}) (map[string]interface{}, error) {
	paymentIdentifier := data["identifier"].(string)
	return t.CancelPayment(paymentIdentifier)
}
