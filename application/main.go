package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"context"
	"time"

	"net/http"

	"bytes"

	//qrcode "github.com/yeqown/go-qrcode" // 给后面的包一个简称

	"github.com/gin-gonic/gin"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	qrcode "github.com/yeqown/go-qrcode"
)

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

func main() {
	router := gin.Default()

	// 定义路由 http://175.24.90.252:8080
	{
		router.POST("/user", userRegister)
		router.POST("/login", logIn)
		router.POST("/help", applyforHelp)
		router.GET("/manager", queryUncheckedHelp)
		router.POST("/manager", checkHelp)
		router.GET("/help/all/:id", queryAllMyHelp)
		router.GET("/allhelp", queryAllHelp)
		router.GET("/support/all/:id", queryAllMySupport)
		router.POST("/support", supportTheHelp)
		router.POST("/support/logi", addSupportLogi)
		router.POST("/support/path", addSupportPath)
		//router.GET("/support/path") todo 手机访问的html的url
		router.GET("/support/qrcode/:id", getQRcode)
		router.GET("/help/support/:uid/:hid", queryAllSupportToTheHelp)
		router.POST("/adopt", adoptSupportToTheHelp)
		router.POST("/ipfs/up", loadPicture)
		router.POST("/ipfs/down", downPicture)

	}

	router.Run() // listen and serve on 8080
}

var sh *shell.Shell

func loadPicture(c *gin.Context) {
	form, _ := c.MultipartForm()
	allfiles := form.File["upload[]"]
	sh = shell.NewShell("localhost:5001")
	allhash := make([]string, 0)
	i := 0
	for _, file := range allfiles {
		fmt.Println(file.Filename)
		s := strconv.Itoa(i)
		i = i + 1
		path := "../images/upphoto" + s + ".jpg"
		c.SaveUploadedFile(file, path)
		f, err := os.Open(path)
		if err != nil {
			return
		}
		hash, err := sh.Add(f)
		if err != nil {
			fmt.Println("上传ipfs时错误：", err)
		}
		allhash = append(allhash, hash)
	}
	allhashbytes, _ := json.Marshal(allhash)
	c.String(http.StatusOK, string(allhashbytes))
}

type downRequest struct {
	Namelist []string `form:"namelist" binding:"required"`
}

func downPicture(c *gin.Context) {
	req := new(downRequest)
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(400, err)
		return
	}
	sh = shell.NewShell("localhost:5001")
	i := 0
	for _, name := range req.Namelist {
		read, err := sh.Cat(name)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer read.Close()
		data, _ := ioutil.ReadAll(read)
		s := strconv.Itoa(i)
		path := "../images/photo" + s + ".jpg"
		i = i + 1
		if ioutil.WriteFile(path, data, 0644) != nil {
			return
		}
		c.FileAttachment(path, "qrcode")
	}

	c.JSON(http.StatusOK, gin.H{"status": "all downloaded!"})
}

type UserRegisterRequest struct {
	Name     string `form:"name" binding:"required"`
	Id       string `form:"id" binding:"required"`
	Phone    string `form:"phone" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// 用户开户
func userRegister(ctx *gin.Context) {
	// 参数处理
	req := new(UserRegisterRequest)
	if err := ctx.ShouldBind(req); err != nil {
		ctx.AbortWithError(400, err)
		return
	}

	// 区块链交互
	resp, err := channelExecute("userRegister", [][]byte{
		[]byte(req.Name),
		[]byte(req.Id),
		[]byte(req.Phone),
		[]byte(req.Password),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

type logInRequest struct {
	Id       string `form:"id" binding:"required"`
	Password string `form:"password" binding:"required"`
}

// 用户登录
func logIn(ctx *gin.Context) {
	// 参数处理
	req := new(logInRequest)
	if err := ctx.ShouldBind(req); err != nil {
		ctx.AbortWithError(400, err)
		return
	}

	// 区块链交互
	resp, err := channelExecute("LogIn", [][]byte{
		[]byte(req.Id),
		[]byte(req.Password),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

type helpRequest struct {
	Id        string      `form:"id" binding:"required"`
	Needlist  []Materials `form:"needlist" binding:"required"`
	Org       string      `form:"org" binding:"required"`
	Address   string      `form:"address" binding:"required"`
	Photolist string      `form:"photolist" binding:"required"`
}

func applyforHelp(ctx *gin.Context) {
	// 参数处理
	req := new(helpRequest)
	if err := ctx.ShouldBind(req); err != nil {
		ctx.AbortWithError(400, err)
		return
	}

	// 区块链交互
	resp, err := channelExecute("applyforHelp", [][]byte{
		[]byte(req.Id),
		[]byte(req.Needlist),
		[]byte(req.Org),
		[]byte(req.Address),
		[]byte(req.Photolist),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

func queryUncheckedHelp(ctx *gin.Context) {
	resp, err := channelQuery("queryUncheckedHelp", [][]byte{})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	//ctx.JSON(http.StatusOK, resp)
	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

type checkRequest struct {
	HId     string `form:"hid" binding:"required"`
	Adopted string `form:"adopted" binding:"required"`
}

func checkHelp(ctx *gin.Context) {
	// 参数处理
	req := new(checkRequest)
	if err := ctx.ShouldBind(req); err != nil {
		ctx.AbortWithError(400, err)
		return
	}

	// 区块链交互
	resp, err := channelExecute("checkHelp", [][]byte{
		[]byte(req.HId),
		[]byte(req.Adopted),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
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

type Photoitem struct {
	Src  string `json:"src"`
	S    string `json:"s"`
	Size string `json:"size"`
}

type ResHelp struct {
	HId        string      `json:"hid"`
	Info       UserInfo    `json:"info"`
	MateArr    []Materials `json:"matearr"`
	Org        string      `json:"org"`
	Address    string      `json:"address"`
	Photoitems []Photoitem `json:"photoitems"`
	Checked    bool        `json:"checked"`
	Adopted    bool        `json:"adopted"`
	Sidlist    []string    `json:"sidlist"`
}

func queryAllMyHelp(ctx *gin.Context) {
	userId := ctx.Param("id")

	resp, err := channelQuery("queryAllMyHelp", [][]byte{
		[]byte(userId),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	//ctx.JSON(http.StatusOK, resp)
	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

type TransferStation struct {
	Location string   `json:"location"`
	Photo    []string `json:"photo"`
	Time     string   `json:"time"`
}

type ResTrans struct {
	Location   string      `json:"location"`
	Photoitems []Photoitem `json:"photoitems"`
	Time       string      `json:"time"`
}

type ResLogistics struct {
	Start       string     `json:"start"`
	End         string     `json:"end"`
	Company     string     `json:"company"`
	TrackingNum string     `json:"trackingnum"`
	ResPath     []ResTrans `json:"respath"`
}

type Logistics struct {
	Start       string            `json:"start"`
	End         string            `json:"end"`
	Company     string            `json:"company"`
	TrackingNum string            `json:"trackingnum"`
	Path        []TransferStation `json:"path"`
}

type ResSupport struct {
	Hid                string       `json:"hid"`
	Hinfo              UserInfo     `json:"hinfo"`
	Sid                string       `json:"sid"`
	Sinfo              UserInfo     `json:"sinfo"`
	Smatearr           []Materials  `json:"smatearr"`
	MaterPhotoitems    []Photoitem  `json:"materphotoitems"`
	ResLogi            ResLogistics `json:"logi"`
	FeedbackPhotoitems []Photoitem  `json:"feedbackphotoitems"`
	FeedbackMessage    string       `json:"feedbackmessage"`
	Finished           bool         `json:"finished"`
}

type Support struct {
	Hid             string      `json:"hid"`
	Hinfo           UserInfo    `json:"hinfo"`
	Sid             string      `json:"sid"`
	Sinfo           UserInfo    `json:"sinfo"`
	Smatearr        []Materials `json:"smatearr"`
	MaterPhoto      []string    `json:"materphoto"`
	Logi            Logistics   `json:"logi"`
	FeedbackPhoto   []string    `json:"feedbackphoto"`
	FeedbackMessage string      `json:"feedbackmessage"`
	Finished        bool        `json:"finished"`
}

func queryAllHelp(ctx *gin.Context) {
	resp, err := channelQuery("queryAllHelp", [][]byte{})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}
	var helps []Help
	_ = json.Unmarshal([]byte(bytes.NewBuffer(resp.Payload).String()), &helps)
	var resHelps []ResHelp
	for _, k := range helps {
		var photoitems []Photoitem
		for _, value := range k.Photo {
			src, s, size := downPicture(value)
			photoitem := Photoitem{
				Src:  src,
				S:    s,
				Size: size,
			}
			photoitems = append(photoitems, photoitem)
		}
		resHelp := ResHelp{
			HId:        k.HId,
			Info:       k.Info,
			MateArr:    k.MateArr,
			Org:        k.Org,
			Address:    k.Address,
			Photoitems: photoitems,
			Checked:    k.Checked,
			Adopted:    k.Adopted,
			Sidlist:    k.Sidlist,
		}
		resHelps = append(resHelps, resHelp)
	}
	reshelpsBytes, _ := json.Marshal(resHelps)
	ctx.String(http.StatusOK, string(reshelpsBytes))
}

func queryAllMySupport(ctx *gin.Context) {
	userId := ctx.Param("uid")
	resp, err := channelQuery("queryAllMySupport", [][]byte{
		[]byte(userId),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	//ctx.JSON(http.StatusOK, resp)
	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

type supportRequest struct {
	HId         string `form:"hid" binding:"required"`
	HUID        string `form:"huid" binding:"required"`
	SUID        string `form:"suid" binding:"required"`
	Supportlist string `form:"supportlist" binding:"required"`
	Photolist   string `form:"photolist" binding:"required"`
}

func supportTheHelp(ctx *gin.Context) {
	req := new(supportRequest)
	if err := ctx.ShouldBind(req); err != nil {
		ctx.AbortWithError(400, err)
		return
	}

	// 区块链交互
	resp, err := channelExecute("supportTheHelp", [][]byte{
		[]byte(req.HId),
		[]byte(req.HUID),
		[]byte(req.SUID),
		[]byte(req.Supportlist),
		[]byte(req.Photolist),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

type addSupportLogiRequest struct {
	Id          string `form:"id" binding:"required"`
	Start       string `form:"start" binding:"required"`
	End         string `form:"end" binding:"required"`
	Company     string `form:"company" binding:"required"`
	Trackingnum string `form:"tackingnum" binding:"required"`
}

func addSupportLogi(ctx *gin.Context) {
	req := new(addSupportLogiRequest)
	if err := ctx.ShouldBind(req); err != nil {
		ctx.AbortWithError(400, err)
		return
	}

	// 区块链交互
	resp, err := channelExecute("addSupportLogi", [][]byte{
		[]byte(req.Id),
		[]byte(req.Start),
		[]byte(req.End),
		[]byte(req.Company),
		[]byte(req.Trackingnum),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

type addSupportPathRequest struct {
	Id        string `form:"id"        binding:"required"`
	Location  string `form:"location"  binding:"required"`
	Photolist string `form:"photolist" binding:"required"`
	Time      string `form:"time"      binding:"required"`
}

func addSupportPath(ctx *gin.Context) {
	req := new(addSupportPathRequest)
	if err := ctx.ShouldBind(req); err != nil {
		ctx.AbortWithError(400, err)
		return
	}

	// 区块链交互
	resp, err := channelExecute("addSupportPath", [][]byte{
		[]byte(req.Id),
		[]byte(req.Location),
		[]byte(req.Photolist),
		[]byte(req.Time),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

func queryAllSupportToTheHelp(ctx *gin.Context) {
	UId := ctx.Param("uid")
	HId := ctx.Param("hid")
	resp, err := channelQuery("queryAllSupportToTheHelp", [][]byte{
		[]byte(UId),
		[]byte(HId),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	//ctx.JSON(http.StatusOK, resp)
	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

type adoptedRequest struct {
	SId       string `form:"sid"       binding:"required"`
	Photolist string `form:"photolist" binding:"required"`
	Message   string `form:"message"   binding:"required"`
}

func adoptSupportToTheHelp(ctx *gin.Context) {
	req := new(adoptedRequest)
	if err := ctx.ShouldBind(req); err != nil {
		ctx.AbortWithError(400, err)
		return
	}

	// 区块链交互
	resp, err := channelExecute("adoptSupportToTheHelp", [][]byte{
		[]byte(req.SId),
		[]byte(req.Photolist),
		[]byte(req.Message),
	})

	if err != nil {
		ctx.String(http.StatusOK, err.Error())
		return
	}

	ctx.String(http.StatusOK, bytes.NewBuffer(resp.Payload).String())
}

func getQRcode(ctx *gin.Context) {

	id := ctx.Param("id")
	url := "http://175.24.90.252:8080" //todo
	qrc, err := qrcode.New(url)
	if err != nil {
		fmt.Printf("could not generate QRCode: %v", err)
	}
	// 保存二维码
	if err := qrc.Save("../images/images.jpg"); err != nil {
		fmt.Printf("could not save image: %v", err)
	}
	//todo 发送出去
	ctx.FileAttachment("../images/images.jpg", "qrcode")
	ctx.JSON(http.StatusOK, gin.H{"status": "qrcode downloaded!"})
}

var (
	sdk           *fabsdk.FabricSDK
	channelName   = "mychannel"
	chaincodeName = "charity1"
	org           = "org1"
	user          = "Admin"
	//configPath    = "$GOPATH/src/github.com/hyperledger/fabric/imocc/application/config.yaml"
	configPath = "./config.yaml"
)

func init() {
	var err error
	sdk, err = fabsdk.New(config.FromFile(configPath))
	if err != nil {
		panic(err)
	}
}

// 区块链管理
func manageBlockchain() {
	// 表明身份,说明自己要处理哪一个channel 自己属于哪个组织 操作哪个用户
	ctx := sdk.Context(fabsdk.WithOrg(org), fabsdk.WithUser(user))

	cli, err := resmgmt.New(ctx)
	if err != nil {
		panic(err)
	}

	// 具体操作
	cli.SaveChannel(resmgmt.SaveChannelRequest{}, resmgmt.WithOrdererEndpoint("orderer.imocc.com"), resmgmt.WithTargetEndpoints())
}

// 区块链数据查询 账本的查询
func queryBlockchain() {
	ctx := sdk.ChannelContext(channelName, fabsdk.WithOrg(org), fabsdk.WithUser(user))

	cli, err := ledger.New(ctx)
	if err != nil {
		panic(err)
	}

	resp, err := cli.QueryInfo(ledger.WithTargetEndpoints("peer0.org1.imocc.com"))
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)

	// 1
	cli.QueryBlockByHash(resp.BCI.CurrentBlockHash)

	// 2
	for i := uint64(0); i <= resp.BCI.Height; i++ {
		cli.QueryBlock(i)
	}
}

// 区块链交互
func channelExecute(fcn string, args [][]byte) (channel.Response, error) {
	ctx := sdk.ChannelContext(channelName, fabsdk.WithOrg(org), fabsdk.WithUser(user))

	cli, err := channel.New(ctx)
	if err != nil {
		return channel.Response{}, err
	}

	// 状态更新，insert/update/delete
	resp, err := cli.Execute(channel.Request{
		ChaincodeID: chaincodeName,
		Fcn:         fcn,
		Args:        args,
	}, channel.WithTargetEndpoints("peer0.org1.imocc.com"))
	if err != nil {
		return channel.Response{}, err
	}

	// 链码事件监听，这里不必要
	go func() {
		// channel
		reg, ccevt, err := cli.RegisterChaincodeEvent(chaincodeName, "eventname")
		if err != nil {
			return
		}
		defer cli.UnregisterChaincodeEvent(reg)

		timeoutctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		for {
			select {
			case evt := <-ccevt:
				fmt.Printf("received event of tx %s: %+v", resp.TransactionID, evt)
			case <-timeoutctx.Done():
				fmt.Println("event timeout, exit!")
				return
			}
		}

		// event
		//eventcli, err := event.New(ctx)
		//if err != nil {
		//	return
		//}

		//eventcli.RegisterChaincodeEvent(chaincodeName, "eventname")
	}()

	// 交易状态事件监听
	go func() {
		eventcli, err := event.New(ctx)
		if err != nil {
			return
		}

		reg, status, err := eventcli.RegisterTxStatusEvent(string(resp.TransactionID))
		defer eventcli.Unregister(reg) // 注册必有注销

		timeoutctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		for {
			select {
			case evt := <-status:
				fmt.Printf("received event of tx %s: %+v", resp.TransactionID, evt)
			case <-timeoutctx.Done(): //防止死等超时
				fmt.Println("event timeout, exit!")
				return
			}
		}
	}()

	return resp, nil
}

func channelQuery(fcn string, args [][]byte) (channel.Response, error) {
	ctx := sdk.ChannelContext(channelName, fabsdk.WithOrg(org), fabsdk.WithUser(user))

	cli, err := channel.New(ctx)
	if err != nil {
		return channel.Response{}, err
	}

	// 状态的查询，select
	return cli.Query(channel.Request{
		ChaincodeID: chaincodeName,
		Fcn:         fcn,
		Args:        args,
	}, channel.WithTargetEndpoints("peer0.org1.imocc.com"))
}

// 事件监听
func eventHandle() {
	ctx := sdk.ChannelContext(channelName, fabsdk.WithOrg(org), fabsdk.WithUser(user))

	cli, err := event.New(ctx)
	if err != nil {
		panic(err)
	}

	// 交易状态事件
	// 链码事件 业务事件
	// 区块事件
	reg, blkevent, err := cli.RegisterBlockEvent()
	if err != nil {
		panic(err)
	}
	defer cli.Unregister(reg)

	timeoutctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for {
		select {
		case evt := <-blkevent:
			fmt.Printf("received a block", evt)
		case <-timeoutctx.Done():
			fmt.Println("event timeout, exit!")
			return
		}
	}
}
