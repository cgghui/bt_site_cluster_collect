package collect

import (
	"errors"
	"github.com/cgghui/cgghui"
	"net/http"
	"strconv"
	"time"
)

type Tag uint8

const (
	TagCommerce   Tag = 0  // 电商
	TagMobile     Tag = 1  // 手机
	TagCar        Tag = 2  // 汽车
	TagSmart      Tag = 3  // 智能 （包括：无人机 VR 机器人 人工智能）
	TagIT         Tag = 4  // IT (包括：百度 大数据 谷歌 阿里巴巴...)
	TagTX         Tag = 5  // 通讯 (包括：手机 苹果 华为 小米...)
	TagLife       Tag = 6  // 生活 (包括：电商 支付 直播 共享单车...)
	TagSAB        Tag = 7  // 创业
	TagScience    Tag = 8  // 科学
	TagDigital    Tag = 9  // 数码
	TagFashion    Tag = 10 // 时尚
	TagInternet   Tag = 11 // 互联网
	TagBlockChain Tag = 12 // 区块链
	Tag5G         Tag = 13 // 5G
	TagBaBy       Tag = 14 // 亲子
	TagArt        Tag = 15 // 艺术
	TagMobileTest Tag = 16 // 手机评测
	TagTravel     Tag = 17 // 美食
)

const (
	TagClass     = ".tag"
	TagAttrName  = "data-name"
	TagAttrValue = "data-tag"
)

const UserAgentChrome = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.75 Safari/537.36"
const UserAgentSogouSpider = "Sogou web spider/4.0(+http://www.sogou.com/docs/help/webmasters.htm#07)"
const UserAgentBaiduSpider = "Mozilla/5.0 (compatible; Baiduspider-render/2.0; +http://www.baidu.com/search/spider.html)"

var HttpClient = &http.Client{
	Timeout: 6 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

var ErrUndefinedTag = errors.New("undefined tag")
var ErrUndefinedArticleHref = errors.New("undefined article href")
var ErrArticleTooShort = errors.New("article too short")

// RequestStructure 构造请求
func RequestStructure(req *http.Request, spider ...bool) {
	if len(spider) > 0 && spider[0] {
		req.Header.Add("User-Agent", UserAgentBaiduSpider)
		req.Header.Add("X-Forwarded-For", cgghui.RandomSliceString(&baiduSpiderIP)+strconv.FormatInt(cgghui.RangeRand(1, 255), 10))
	} else {
		req.Header.Add("User-Agent", UserAgentChrome)
	}
}

var baiduSpiderIP = []string{"116.179.37.", "124.166.232.", "116.179.32.", "180.76.15.", "180.76.5."}

type ArticleTag struct {
	Name string
	Tag  string
}

// Article 文章
type Article struct {
	Title       string       // 标题
	Content     string       // 正文
	Alias       string       // 别名
	Tag         []ArticleTag // 标签
	Cate        Category     // 分类
	AuthorName  string       // 作者
	PostTime    time.Time    // 发布时间
	Intro       string       // 摘要
	Href        string       // 链接
	LocalImages []string     // 本地下载的图片
}

// Category 分类
type Category struct {
	Name     string // 名称
	Alias    string // 别名
	Order    string // 排序
	ParentID int    // 父级
	Intro    string // 简述
}
