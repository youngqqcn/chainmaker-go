/*
   Created by guoxin in 2022/2/28 11:49 AM
*/
package tx_duplicate_function

import (
	"chainmaker.org/chainmaker/common/v2/random/uuid"
	sdkutils "chainmaker.org/chainmaker/sdk-go/v2/utils"
	"log"
	"strconv"
	"testing"
	"time"
)

var (
	// Solo
	caPaths = [][]string{
		{"../../config/crypto-config/wx-org1.chainmaker.org/ca"},
	}
	userKeyPaths = []string{
		"../../config/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.key",
	}
	userCrtPaths = []string{
		"../../config/crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.crt",
	}
	IPs = []string{
		"127.0.0.1",
	}
	// 四节点
	//caPaths = [][]string{
	//	{certPathPrefix + "crypto-config/wx-org1.chainmaker.org/ca"},
	//	{certPathPrefix + "crypto-config/wx-org2.chainmaker.org/ca"},
	//	{certPathPrefix + "crypto-config/wx-org3.chainmaker.org/ca"},
	//	{certPathPrefix + "crypto-config/wx-org4.chainmaker.org/ca"},
	//}
	//userKeyPaths = []string{
	//	certPathPrefix + "crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.key",
	//	certPathPrefix + "crypto-config/wx-org2.chainmaker.org/user/client1/client1.sign.key",
	//	certPathPrefix + "crypto-config/wx-org3.chainmaker.org/user/client1/client1.sign.key",
	//	certPathPrefix + "crypto-config/wx-org4.chainmaker.org/user/client1/client1.sign.key",
	//}
	//userCrtPaths = []string{
	//	certPathPrefix + "crypto-config/wx-org1.chainmaker.org/user/client1/client1.sign.crt",
	//	certPathPrefix + "crypto-config/wx-org2.chainmaker.org/user/client1/client1.sign.crt",
	//	certPathPrefix + "crypto-config/wx-org3.chainmaker.org/user/client1/client1.sign.crt",
	//	certPathPrefix + "crypto-config/wx-org4.chainmaker.org/user/client1/client1.sign.crt",
	//}
	//orgIds = []string{
	//	"wx-org1.chainmaker.org",
	//	"wx-org2.chainmaker.org",
	//	"wx-org3.chainmaker.org",
	//	"wx-org4.chainmaker.org",
	//}
	//IPs = []string{
	//	"127.0.0.1",
	//	"127.0.0.1",
	//	"127.0.0.1",
	//	"127.0.0.1",
	//}

	Ports = []int{
		12301,
		12302,
		12303,
		12304,
	}

	Size = 5

	IndexTimestampTx    = 1
	IndexNotTimestampTx = 2
)

func TestDuplicateTimestampTx(t *testing.T) {
	txIds := make([]string, Size)
	for i := 0; i < Size; i++ {
		txIds[i] = strconv.FormatInt(time.Now().UnixNano()/1e6, 10) + "-" + uuid.GetUUID()
		log.Println(txIds[i])
	}
	if Exist(IndexTimestampTx) {
		Delete(IndexTimestampTx)
	}
	WriteSlices(txIds, IndexTimestampTx)
	for i := 0; i < Size; i++ {
		SendTx(txIds[i], userCrtPaths[0], userKeyPaths[0], IPs[0], Ports[0], caPaths[0])
	}
	for i := 0; i < Size; i++ {
		log.Println(txIds[i])
		SendTx(txIds[i], userCrtPaths[0], userKeyPaths[0], IPs[0], Ports[0], caPaths[0])
	}
}

func TestDuplicateTimestampTxRestartExists(t *testing.T) {
	txIds := ReadSlices(IndexTimestampTx)
	for i := 0; i < len(txIds); i++ {
		log.Println(txIds[i])
		SendTx(txIds[i], userCrtPaths[0], userKeyPaths[0], IPs[0], Ports[0], caPaths[0])
	}
}
func TestDuplicateTx(t *testing.T) {
	txIds := make([]string, Size)
	for i := 0; i < Size; i++ {
		txIds[i] = sdkutils.GetRandTxId()
		log.Println(txIds[i])
	}
	if Exist(IndexNotTimestampTx) {
		Delete(IndexNotTimestampTx)
	}
	WriteSlices(txIds, IndexNotTimestampTx)
	for i := 0; i < Size; i++ {
		SendTx(txIds[i], userCrtPaths[0], userKeyPaths[0], IPs[0], Ports[0], caPaths[0])
	}
	for i := 0; i < Size; i++ {
		log.Println(txIds[i])
		SendTx(txIds[i], userCrtPaths[0], userKeyPaths[0], IPs[0], Ports[0], caPaths[0])
	}
}

func TestDuplicateTxRestartExists(t *testing.T) {
	txIds := ReadSlices(IndexNotTimestampTx)
	for i := 0; i < len(txIds); i++ {
		log.Println(txIds[i])
		SendTx(txIds[i], userCrtPaths[0], userKeyPaths[0], IPs[0], Ports[0], caPaths[0])
	}
}
