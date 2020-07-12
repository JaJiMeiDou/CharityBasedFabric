package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Charity   .
type Charity struct{}

type User struct {
	Name     string `json:"name"`
	Id       string `json:"id"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Hnum     string `json:"hnum"`
	Snum     string `json:"snum"`
}

type UserInfo struct {
	Name  string `json:"name"`
	Id    string `json:"id"`
	Phone string `json:"phone"`
}

type Materials struct {
	Category string `json:"category"`
	Number   int    `json:"number"`
}

type TransferStation struct {
	Location string   `json:"location"`
	Photo    []string `json:"photo"`
	Time     string   `json:"time"`
}

type Logistics struct {
	Start       string            `json:"start"`
	End         string            `json:"end"`
	Company     string            `json:"company"`
	TrackingNum string            `json:"trackingnum"`
	Path        []TransferStation `json:"path"`
}

type Help struct {
	HId     string      `json:"hid"`
	Info    UserInfo    `json:"info"`
	MateArr []Materials `json:"matearr"`
	Org     string      `json:"org"`
	Address string      `json:"address"`
	Photo   []string    `json:"photo"`
	Checked bool        `json:"checked"`
	Adopted bool        `json:"adopted"`
	Sidlist []string    `json:"sidlist"`
}

type Support struct {
	Hid             string      `json:"hid"`
	Hinfo           UserInfo    `json:"hinfo"`
	Sid             string      `json:"sid"`
	Sinfo           UserInfo    `json:"sinfo"`
	Smatearr        []Materials `json:"smatearr"`
	MaterPhoto      []string    `json:"materphoto"`
	Logi            Logistics   `json:"logi"`
	FeedbackPhoto   []string      `json:"feedbackphoto"`
	FeedbackMessage string      `json:"feedbackmessage"`
	Finished        bool        `json:"finished"`
}

func constructUserKey(userId string) string {
	return fmt.Sprintf("user_%s", userId)
}
func constructHelpKey(HId string) string {
	return fmt.Sprintf("help_%s", HId)
}
func constructSupportKey(SID string) string {
	return fmt.Sprintf("Support_%s", SID)
}

func userRegister(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 4 {
		return shim.Error("not enough args")
	}
	name := args[0]
	id := args[1]
	phone := args[2]
	password := args[3]
	snum := "0"
	hnum := "0"

	if name == "" || id == "" || phone == "" || password == "" {
		return shim.Error("invalid args")
	}

	if userBytes, err := stub.GetState(constructUserKey(id)); err == nil && len(userBytes) != 0 {
		return shim.Error("user already exist")
	}

	user := &User{
		Name:     name,
		Id:       id,
		Phone:    phone,
		Password: password,
		Hnum:     hnum,
		Snum:     snum,
	}
	//序列化并存储
	userBytes, err := json.Marshal(user)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal user error %s", err))
	}

	if err := stub.PutState(constructUserKey(id), userBytes); err != nil {
		return shim.Error(fmt.Sprintf("put user error %s", err))
	}

	return shim.Success(nil)
}

func logIn(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("not enough args")
	}
	id := args[0]
	password := args[1]

	if id == "" || password == "" {
		return shim.Error("invalid args")
	}

	userBytes, err := stub.GetState(constructUserKey(id))
	if err == nil && len(userBytes) == 0 {
		return shim.Error("no such user")
	}

	user := new(User)
	//反序列化
	if err := json.Unmarshal(userBytes, user); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal user error: %s", err))
	}

	if password != user.Password {
		return shim.Error("error password")
	}

	user.Password = ""
	userBytes, err = json.Marshal(user)
	return shim.Success(userBytes)
}

func applyforHelp(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 5 {
		return shim.Error("not enough args")
	}
	id := args[0]
	needlistBytes := []byte(args[1])
	org := args[2]
	address := args[3]
	photoBytes := []byte(args[4])
	if id == "" || needlistBytes == nil || org == "" || address == "" || photoBytes == nil {
		return shim.Error("invalid args")
	}
	// 该用户是否存在
	userBytes, err := stub.GetState(constructUserKey(id))
	if err == nil && len(userBytes) == 0 {
		return shim.Error("user dont exist")
	}

	var needlist []Materials
	if err := json.Unmarshal(needlistBytes, &needlist); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal needlistBytes error: %s", err))
	}
	var photolist []string
	if err := json.Unmarshal(photoBytes, &photolist); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal photoBytes error: %s", err))
	}
	// 获得该用户的hunm以构成hid
	user := new(User)
	//反序列化到user中
	if err := json.Unmarshal(userBytes, user); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal user error: %s", err))
	}

	hid := id + user.Hnum

	info := UserInfo{
		Name:  user.Name,
		Id:    id,
		Phone: user.Phone,
	}

	help := &Help{
		HId:     hid,
		Info:    info,
		MateArr: needlist,
		Org:     org,
		Address: address,
		Photo:   photolist,
		Checked: false,
		Adopted: false,
		Sidlist: make([]string, 0),
	}
	// 存help 改user.Hnum
	helpBytes, err := json.Marshal(help)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal help error %s", err))
	}

	i, err := strconv.Atoi(user.Hnum)
	if err != nil {
		return shim.Error(fmt.Sprintf("strconv.Atoi help user.Hnum error %s", err))
	}
	i++
	user.Hnum = strconv.Itoa(i)
	userBytes, err = json.Marshal(user)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal user error %s", err))
	}

	helpKey, err := stub.CreateCompositeKey("Help", []string{"help",id,hid})
	if err != nil {
		return shim.Error(fmt.Sprintf("create key error: %s", err))
	}
	if err := stub.PutState(helpKey, helpBytes); err != nil {
		return shim.Error(fmt.Sprintf("put help error %s", err))
	}

	//if err := stub.PutState(constructHelpKey(hid), helpBytes); err != nil {
	//	return shim.Error(fmt.Sprintf("put help error %s", err))
	//}
	if err := stub.PutState(constructUserKey(id), userBytes); err != nil {
		return shim.Error(fmt.Sprintf("update user error %s", err))
	}

	return shim.Success([]byte(hid))

}

func queryUncheckedHelp(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 0 {
		return shim.Error("too many args")
	}
	//keys := make([]string, 0)
	result, err := stub.GetStateByPartialCompositeKey("Help", []string{"help"})
	if err != nil {
		return shim.Error(fmt.Sprintf("query UncheckedHelp  error: %s", err))
	}
	defer result.Close()

	helps := make([]*Help, 0)
	for result.HasNext() {
		helpVal, err := result.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("query error: %s", err))
		}

		help := new(Help)
		if err := json.Unmarshal(helpVal.GetValue(), help); err != nil {
			return shim.Error(fmt.Sprintf("unmarshal error: %s", err))
		}

		// 过滤掉已核对的记录
		if help.Checked == true {
			continue
		}

		helps = append(helps, help)
	}
	helpsBytes, err := json.Marshal(helps)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal error: %s", err))
	}

	return shim.Success(helpsBytes)
}

func checkHelp(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("too many args")
	}
	hid := args[0]
	var adopted bool
	if args[1] == "true" {
		adopted = true
	}else{
		adopted = false
	}
	uid := args[2]
	if hid == "" || (args[1] != "true" && args[1] != "false")|| uid == "" {
		return shim.Error("invalid args")
	}
	helpKey, err := stub.CreateCompositeKey("Help", []string{"help",uid,hid})
	if err != nil {
		return shim.Error(fmt.Sprintf("create key error: %s", err))
	}
	helpBytes, err := stub.GetState(helpKey)
	if err != nil || len(helpBytes) == 0 {
		return shim.Error("help not found")
	}
	help := new(Help)
	//反序列化
	if err := json.Unmarshal(helpBytes, help); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal help error: %s", err))
	}
	help.Checked = true
	help.Adopted = adopted

	helpBytes, err = json.Marshal(help)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal help error %s", err))
	}
	
	if err := stub.PutState(helpKey, helpBytes); err != nil {
		return shim.Error(fmt.Sprintf("put help error %s", err))
	}
	return shim.Success(nil)
}

func queryAllMyHelp(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("args num fault")
	}
	uid := args[0]
	if uid == "" {
		return shim.Error("invalid args")
	}
	// 该用户是否存在
	userBytes, err := stub.GetState(constructUserKey(uid)) 
	if err == nil && len(userBytes) == 0 {
		return shim.Error("user dont exist")
	}
	user := new(User)
	//反序列化到user中
	if err := json.Unmarshal(userBytes, user); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal user error: %s", err))
	}
	num := user.Hnum
	helps := make([]*Help, 0)
	if num != "0" {
		result, err := stub.GetStateByPartialCompositeKey("Help", []string{"help",uid})
		if err != nil {
			return shim.Error(fmt.Sprintf("query AllMyHelp  error: %s", err))
		}
		defer result.Close()
		for result.HasNext() {
			helpVal, err := result.Next()
			if err != nil {
				return shim.Error(fmt.Sprintf("query error: %s", err))
			}

			help := new(Help)
			if err := json.Unmarshal(helpVal.GetValue(), help); err != nil {
				return shim.Error(fmt.Sprintf("unmarshal error: %s", err))
			}

			helps = append(helps, help)
		}
	}
	helpsBytes, err := json.Marshal(helps)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal error: %s", err))
	}

	return shim.Success(helpsBytes)
}

func queryAllHelp(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 0 {
		return shim.Error("too many args")
	}
	result, err := stub.GetStateByPartialCompositeKey("Help", []string{"help"})
	if err != nil {
		return shim.Error(fmt.Sprintf("query AllHelp  error: %s", err))
	}
	defer result.Close()

	helps := make([]*Help, 0)
	for result.HasNext() {
		helpVal, err := result.Next()
		if err != nil {
			return shim.Error(fmt.Sprintf("query error: %s", err))
		}

		help := new(Help)
		if err := json.Unmarshal(helpVal.GetValue(), help); err != nil {
			return shim.Error(fmt.Sprintf("unmarshal error: %s", err))
		}

		// 过滤掉未审核或者审核不通过的记录
		if help.Checked == false || (help.Checked == true && help.Adopted == false) {
			continue
		}

		helps = append(helps, help)
	}
	helpsBytes, err := json.Marshal(helps)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal error: %s", err))
	}

	return shim.Success(helpsBytes)
}

func queryAllMySupport(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("args num fault")
	}
	uid := args[0]
	if uid == "" {
		return shim.Error("invalid args")
	}
	// 该用户是否存在
	userBytes, err := stub.GetState(constructUserKey(uid))
	if err == nil && len(userBytes) == 0 {
		return shim.Error("user dont exist")
	}
	user := new(User)
	//反序列化到user中
	if err := json.Unmarshal(userBytes, user); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal user error: %s", err))
	}
	num := user.Snum
	supports := make([]*Support, 0)
	if num != "0" {
		start := "Support_" + uid + "0"
		end := "Support_" + uid + num
		result, err := stub.GetStateByRange(start, end)
		if err != nil {
			return shim.Error(fmt.Sprintf("query AllMySupport  error: %s", err))
		}
		defer result.Close()
		for result.HasNext() {
			supportVal, err := result.Next()
			if err != nil {
				return shim.Error(fmt.Sprintf("query error: %s", err))
			}

			support := new(Support)
			if err := json.Unmarshal(supportVal.GetValue(), support); err != nil {
				return shim.Error(fmt.Sprintf("unmarshal error: %s", err))
			}

			supports = append(supports, support)
		}
	}
	supportsBytes, err := json.Marshal(supports)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal error: %s", err))
	}

	return shim.Success(supportsBytes)
}

func supportTheHelp(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 5 {
		return shim.Error("args num fault")
	}
	hid := args[0]
	huid := args[1]
	suid := args[2]
	supportlistBytes := []byte(args[3])
	photoBytes := []byte(args[4])

	if hid == "" || len(supportlistBytes) == 0 || huid == "" || suid == "" || len(photoBytes) == 0 {
		return shim.Error("invalid args")
	}
	// 双方用户是否存在
	huserBytes, err := stub.GetState(constructUserKey(huid))
	if err == nil && len(huserBytes) == 0 {
		return shim.Error("huser dont exist")
	}
	suserBytes, err := stub.GetState(constructUserKey(suid))
	if err == nil && len(suserBytes) == 0 {
		return shim.Error("suser dont exist")
	}
	//求助信息是否存在
	helpKey, _ := stub.CreateCompositeKey("Help", []string{"help",huid,hid})
	helpBytes, err := stub.GetState(helpKey)
	if err == nil && len(helpBytes) == 0 {
		return shim.Error("help dont exist")
	}

	var supportlist []Materials
	if err := json.Unmarshal(supportlistBytes, &supportlist); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal supportlistBytes error: %s", err))
	}
	var photolist []string
	if err := json.Unmarshal(photoBytes, &photolist); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal photoBytes error: %s", err))
	}

	// 获得援助用户的sunm以构成sid
	suser := new(User)
	//反序列化到user中
	if err := json.Unmarshal(suserBytes, suser); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal suser error: %s", err))
	}

	sid := suid + suser.Snum

	sinfo := UserInfo{
		Name:  suser.Name,
		Id:    suid,
		Phone: suser.Phone,
	}

	huser := new(User)
	//反序列化到user中
	if err := json.Unmarshal(huserBytes, huser); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal huser error: %s", err))
	}

	hinfo := UserInfo{
		Name:  huser.Name,
		Id:    huid,
		Phone: huser.Phone,
	}

	var logi Logistics

	support := &Support{
		Hid:             hid,
		Hinfo:           hinfo,
		Sid:             sid,
		Sinfo:           sinfo,
		Smatearr:        supportlist,
		MaterPhoto:      photolist,
		Logi:            logi,
		FeedbackPhoto:   make([]string,0),
		FeedbackMessage: "",
		Finished:        false,
	}
	//存support ，改help的list ，改suser
	supportBytes, err := json.Marshal(support)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal support error %s", err))
	}

	i, err := strconv.Atoi(suser.Snum)
	if err != nil {
		return shim.Error(fmt.Sprintf("strconv.Atoi support suser.Snum error %s", err))
	}
	i++
	suser.Snum = strconv.Itoa(i)
	suserBytes, err = json.Marshal(suser)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal suser error %s", err))
	}

	help := new(Help)
	if err := json.Unmarshal(helpBytes, help); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal help error: %s", err))
	}
	help.Sidlist = append(help.Sidlist, sid)
	helpBytes, err = json.Marshal(help)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal help error %s", err))
	}
	if err := stub.PutState(constructSupportKey(sid), supportBytes); err != nil {
		return shim.Error(fmt.Sprintf("put support error %s", err))
	}
	if err := stub.PutState(helpKey, helpBytes); err != nil {
		return shim.Error(fmt.Sprintf("put help error %s", err))
	}
	if err := stub.PutState(constructUserKey(suid), suserBytes); err != nil {
		return shim.Error(fmt.Sprintf("update user error %s", err))
	}

	return shim.Success([]byte(sid))

}

func addSupportLogi(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 5 {
		return shim.Error("args num fault")
	}
	id := args[0]
	start := args[1]
	end := args[2]
	company := args[3]
	trackingnum := args[4]

	if id == "" || start == "" || end == "" || company == "" || trackingnum == "" {
		return shim.Error("invalid args")
	}

	supportBytes, err := stub.GetState(constructSupportKey(id))
	if err == nil && len(supportBytes) == 0 {
		return shim.Error("support dont exist")
	}

	support := new(Support)
	//反序列化到user中
	if err := json.Unmarshal(supportBytes, support); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal support error: %s", err))
	}

	logi := Logistics{
		Start:       start,
		End:         end,
		Company:     company,
		TrackingNum: trackingnum,
		Path:        make([]TransferStation, 0),
	}

	support.Logi = logi

	supportBytes, err = json.Marshal(support)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal support error %s", err))
	}

	if err := stub.PutState(constructSupportKey(id), supportBytes); err != nil {
		return shim.Error(fmt.Sprintf("put support error %s", err))
	}
	return shim.Success(nil)

}

func addSupportPath(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 4 {
		return shim.Error("args num fault")
	}
	id := args[0]
	location := args[1]
	photoBytes := []byte(args[2])
	time := args[3]
	if id == "" || location == "" || len(photoBytes) == 0 || time == "" {
		return shim.Error("invalid args")
	}

	supportBytes, err := stub.GetState(constructSupportKey(id))
	if err == nil && len(supportBytes) == 0 {
		return shim.Error("support dont exist")
	}

	var photolist []string
	if err := json.Unmarshal(photoBytes, &photolist); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal photoBytes error: %s", err))
	}

	path := TransferStation{
		Location: location,
		Photo:    photolist,
		Time:     time,
	}

	support := new(Support)
	//反序列化到user中
	if err := json.Unmarshal(supportBytes, support); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal support error: %s", err))
	}
	support.Logi.Path = append(support.Logi.Path, path)
	supportBytes, err = json.Marshal(support)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal support error %s", err))
	}

	if err := stub.PutState(constructSupportKey(id), supportBytes); err != nil {
		return shim.Error(fmt.Sprintf("put support error %s", err))
	}
	return shim.Success(nil)
}

func queryAllSupportToTheHelp(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("args num fault")
	}
	uid := args[0]
	hid := args[1]
	if uid == "" || hid == "" {
		return shim.Error("invalid args")
	}
	userBytes, err := stub.GetState(constructUserKey(uid))
	if err == nil && len(userBytes) == 0 {
		return shim.Error("user dont exist")
	}
	helpKey, _ := stub.CreateCompositeKey("Help", []string{"help",uid,hid})
	helpBytes, err := stub.GetState(helpKey)
	if err == nil && len(helpBytes) == 0 {
		return shim.Error("help dont exist")
	}
	help := new(Help)
	if err := json.Unmarshal(helpBytes, help); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal help error: %s", err))
	}
	sidlist := help.Sidlist
	supportlist := make([]*Support, 0)
	for _, v := range sidlist {
		supportBytes, err := stub.GetState(constructSupportKey(v)) 
		if err == nil && len(supportBytes) == 0 {
			return shim.Error("support dont exist")
		}
		support := new(Support)
		if err := json.Unmarshal(supportBytes, support); err != nil {
			return shim.Error(fmt.Sprintf("unmarshal support error: %s", err))
		}
		supportlist = append(supportlist, support)
	}
	supportlistBytes, err := json.Marshal(supportlist)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal supportlist error %s", err))
	}
	return shim.Success(supportlistBytes)
}

func adoptSupportToTheHelp(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("args num fault")
	}
	sid := args[0]
	photolistBytes := []byte(args[1])
	message := args[2]
	if sid == "" || len(photolistBytes) == 0 || message == "" {
		return shim.Error("invalid args")
	}
	supportBytes, err := stub.GetState(constructSupportKey(sid));
	if err == nil && len(supportBytes) == 0 {
		return shim.Error("support dont exist")
	}
	support := new(Support)
	if err := json.Unmarshal(supportBytes, support); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal support error: %s", err))
	}
	var photolist []string
	if err := json.Unmarshal(photolistBytes, &photolist); err != nil {
		return shim.Error(fmt.Sprintf("unmarshal photolistBytes error: %s", err))
	}
	support.FeedbackPhoto = photolist
	support.FeedbackMessage = message
	support.Finished = true
	supportBytes, err = json.Marshal(support)
	if err != nil {
		return shim.Error(fmt.Sprintf("marshal support error %s", err))
	}
	if err := stub.PutState(constructSupportKey(sid), supportBytes); err != nil {
		return shim.Error(fmt.Sprintf("put support error %s", err))
	}
	return shim.Success(nil)
}

// Init is called during Instantiate transaction after the chaincode container
// has been established for the first time, allowing the chaincode to
// initialize its internal data
func (c *Charity) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke is called to update or query the ledger in a proposal transaction.
// Updated state variables are not committed to the ledger until the
// transaction is committed.
func (c *Charity) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	funcName, args := stub.GetFunctionAndParameters()
	switch funcName {
	case "userRegister":
		return userRegister(stub, args)
	case "LogIn":
		return logIn(stub, args)
	case "applyforHelp":
		return applyforHelp(stub, args)
	case "queryUncheckedHelp":
		return queryUncheckedHelp(stub, args)
	case "checkHelp":
		return checkHelp(stub, args)
	case "queryAllMyHelp":
		return queryAllMyHelp(stub, args)
	case "queryAllHelp":
		return queryAllHelp(stub, args)
	case "queryAllMySupport":
		return queryAllMySupport(stub, args)
	case "supportTheHelp":
		return supportTheHelp(stub, args)
	case "addSupportLogi":
		return addSupportLogi(stub, args)
	case "addSupportPath":
		return addSupportPath(stub, args)
	case "queryAllSupportToTheHelp":
		return queryAllSupportToTheHelp(stub, args)
	case "adoptSupportToTheHelp":
		return adoptSupportToTheHelp(stub, args)
	default:
		return shim.Error(fmt.Sprintf("unsupported function: %s", funcName))
	}
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(Charity))
	if err != nil {
		fmt.Printf("Error starting Charity chaincode: %s", err)
	}
}
