package www_163_com

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/cgghui/bt_site_cluster_collect/collect"
	"github.com/cgghui/cgghui"
	"github.com/mozillazg/go-pinyin"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	Name = "www_163_com"
)

func init() {
	collect.RegisterStandard(Name, func() collect.Standard {
		return &CollectGo{HomeURL: "https://www.163.com/"}
	})
}

var pyArg = pinyin.NewArgs()

//  TagCommerce   Tag = 0  // 电商
//	TagMobile     Tag = 1  // 手机
//	TagCar        Tag = 2  // 汽车
//	TagSmart      Tag = 3  // 智能 （包括：无人机 VR 机器人 人工智能）
//	TagIT         Tag = 4  // IT (包括：百度 大数据 谷歌 阿里巴巴...)
//	TagTX         Tag = 5  // 通讯 (包括：手机 苹果 华为 小米...)
//	TagLife       Tag = 6  // 生活 (包括：电商 支付 直播 共享单车...)
//	TagSAB        Tag = 7  // 创业
//	TagScience    Tag = 8  // 科学
//	TagDigital    Tag = 9  // 数码
//	TagFashion    Tag = 10 // 时尚
//	TagInternet   Tag = 11 // 互联网
//	TagBlockChain Tag = 12 // 区块链
//	Tag5G         Tag = 13 // 5G
//	TagBaBy       Tag = 14 // 亲子
//	TagArt        Tag = 15 // 艺术
//	TagMobileTest Tag = 16 // 手机评测
//	TagTravel     Tag = 17 // 美食

var Column = map[collect.Tag]string{
	//collect.TagCommerce: "",
	collect.TagMobile: "https://mobile.163.com/special/index_datalist{page}/",
	//collect.TagCar:        "",
	collect.TagSmart: "https://tech.163.com/special/00097UHL/smart_datalist{page}.js",
	collect.TagIT:    "https://tech.163.com/special/it_2016{page}/",
	collect.TagTX:    "https://tech.163.com/special/tele_2016{page}/",
	//collect.TagLife:       "",
	//collect.TagSAB:        "",
	collect.TagScience:    "https://tech.163.com/special/techscience{page}/",
	collect.TagDigital:    "https://digi.163.com/special/index_datalist{page}/",
	collect.TagFashion:    "https://fashion.163.com/special/002688FE/fashion_datalist{page}.js",
	collect.TagInternet:   "https://tech.163.com/special/internet_2016{page}/",
	collect.TagBlockChain: "https://tech.163.com/special/blockchain_2018{page}/",
	collect.Tag5G:         "https://tech.163.com/special/5g_2019{page}/",
	collect.TagBaBy:       "https://baby.163.com/special/003687OS/newsdata_hot{page}.js",
	collect.TagArt:        "https://art.163.com/special/00999815/art_redian_api{page}.js",
	collect.TagMobileTest: "https://mobile.163.com/special/index_datalist_cpsh{page}/",
	collect.TagTravel:     "https://travel.163.com/special/00067VF5/fooddatas_travel{page}.js",
}

type CollectGo struct {
	HomeURL string
}

func (c CollectGo) Name() string {
	return Name
}

func (c CollectGo) GetTag() []collect.Tag {
	r := make([]collect.Tag, 0)
	for k := range Column {
		r = append(r, k)
	}
	return r
}

type Article struct {
	Title      string       `json:"title"`
	Desc       string       `json:"desc"`
	DocUrl     string       `json:"docurl"`
	CommentUrl string       `json:"commenturl"`
	TieNum     int          `json:"tienum"`
	TLastID    string       `json:"tlastid"`
	Label      string       `json:"label"`
	Keywords   []ArticleTag `json:"keywords"`
	Time       string       `json:"time"`
	NewsType   string       `json:"newstype"`
	ImgUrl     string       `json:"imgurl"`
}

type ArticleTag struct {
	Link string `json:"akey_link"`
	Name string `json:"keyname"`
}

func (c CollectGo) ArticleList(tag collect.Tag, page int) ([]collect.Article, error) {
	if _, ok := Column[tag]; !ok {
		return nil, collect.ErrUndefinedTag
	}
	target := Column[tag]
	if page <= 1 {
		target = strings.ReplaceAll(target, "{page}", "")
	} else {
		target = strings.ReplaceAll(target, "{page}", fmt.Sprintf("_%02d", page))
	}
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	collect.RequestStructure(req, true)
	var resp *http.Response
	if resp, err = collect.HttpClient.Do(req); err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	articles := make([]collect.Article, 0)
	switch tag {
	case collect.TagInternet, collect.TagIT, collect.TagBlockChain, collect.Tag5G, collect.TagTX, collect.TagScience:
		{
			var doc *goquery.Document
			if doc, err = goquery.NewDocumentFromReader(resp.Body); err != nil {
				return nil, err
			}
			doc.Find("#news-flow-content .bigsize").Each(func(_ int, h3 *goquery.Selection) {
				a := h3.Find("a")
				articles = append(articles, collect.Article{
					Title: a.Text(),
					Href:  a.AttrOr("href", ""),
					Tag:   make([]collect.ArticleTag, 0),
				})
			})
		}
	default:
		{
			b, _ := ioutil.ReadAll(resp.Body)
			b = bytes.Replace(b, []byte("data_callback("), []byte{}, 1)
			b = bytes.TrimRight(b, ")")
			var ret []Article
			if err = json.Unmarshal(b, &ret); err != nil {
				return nil, err
			}
			for _, r := range ret {
				if r.NewsType != "article" {
					continue
				}
				art := collect.Article{
					Title: r.Title,
					Href:  r.DocUrl,
					Tag:   make([]collect.ArticleTag, 0),
				}
				for _, tg := range r.Keywords {
					py := strings.Join(pinyin.LazyPinyin(tg.Name, pyArg), "")
					if py == "" {
						py = tg.Name
					}
					art.Tag = append(art.Tag, collect.ArticleTag{Name: tg.Name, Tag: py})
				}
				articles = append(articles, art)
			}
		}
	}

	return articles, nil
}

func (c CollectGo) HasSnapshot(art *collect.Article) bool {
	if art.Href == "" {
		return false
	}
	dir := cgghui.MD5(art.Href)
	return collect.PathExists("./snapshot/" + Name + "/" + string(dir[0]) + "/" + dir + ".html")
}

func (c CollectGo) ArticleDetail(art *collect.Article) error {
	var err error
	if art.Href == "" {
		return collect.ErrUndefinedArticleHref
	}
	var cache *os.File
	dir := cgghui.MD5(art.Href)
	snapshotPath := "./snapshot/" + Name + "/" + string(dir[0]) + "/"
	_ = os.MkdirAll(snapshotPath, 0666)
	snapshotPath += dir + ".html"
	if cache, err = os.Open(snapshotPath); err != nil {
		var req *http.Request
		if req, err = http.NewRequest(http.MethodGet, art.Href, nil); err != nil {
			return err
		}
		collect.RequestStructure(req, true)
		var resp *http.Response
		if resp, err = collect.HttpClient.Do(req); err != nil {
			return err
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		if cache, err = os.Create(snapshotPath); err == nil {
			_, _ = io.Copy(cache, resp.Body)
		}
		_, _ = cache.Seek(0, io.SeekStart)
	}
	var doc *goquery.Document
	if doc, err = goquery.NewDocumentFromReader(cache); err != nil {
		return err
	}
	art.Title = doc.Find(`meta[property="og:title"]`).AttrOr("content", "")
	art.PostTime, err = time.Parse("2006-01-02 15:04:05", doc.Find(`#ne_wrap`).AttrOr("data-publishtime", ""))
	art.PostTime = art.PostTime.Local()
	if art.LocalImages == nil {
		art.LocalImages = make([]string, 0)
	}
	word := doc.Find(".post_body")
	//
	if word.Find("video").Length() != 0 {
		return ErrVideoArticle
	}
	// 处理图片
	word.Find("img").Each(func(_ int, img *goquery.Selection) {
		src := img.AttrOr("src", "")
		if src == "" {
			return
		}
		src = html.UnescapeString(src)
		if strings.HasSuffix(src, "logo.png") {
			if img.Parent().Is("div") {
				img.Parent().Remove()
			} else {
				img.Remove()
			}
			return
		}
		var imgPath string
		if imgPath, err = collect.DownloadImage(src); err != nil {
			img.Remove()
			return
		}
		if alt := img.AttrOr("alt", ""); len(alt) == 0 {
			img.RemoveAttr("alt")
		} else {
			if strings.Contains(alt, "http://") {
				img.RemoveAttr("alt")
			}
		}
		img.SetAttr("src", imgPath)
		art.LocalImages = append(art.LocalImages, imgPath)
	})
	//// 处理<a>
	if art.Tag == nil {
		art.Tag = make([]collect.ArticleTag, 0)
	}
	word.Find("a").Each(func(_ int, a *goquery.Selection) {
		aHTML, _ := a.Html()
		aHref := a.AttrOr("href", "")
		if aHref == "" {
			a.Remove()
			return
		}
		href, _ := url.Parse(aHref)
		kw := href.Query().Get("keyword")
		if aHTML == "" {
			a.BeforeHtml(aHTML)
			a.Remove()
		} else {
			py := strings.Join(pinyin.LazyPinyin(kw, pyArg), "")
			if py == "" {
				py = kw
			}
			art.Tag = append(art.Tag, collect.ArticleTag{Name: kw, Tag: py})
			a.RemoveAttr("href")
			a.AddClass(collect.TagClass[1:])
			a.SetAttr(collect.TagAttrName, kw)
			a.SetAttr(collect.TagAttrValue, py)
		}
	})
	//// 处理<p>
	word.Find("p").Each(func(i int, p *goquery.Selection) {
		p.RemoveAttr("id")
		if has := p.HasClass("f_center"); has {
			p.SetAttr("style", "text-align: center;")
			p.RemoveClass("f_center")
		}
	})
	word.Find(".Apple-interchange-newline").Remove()
	for {
		p := word.Find("p")
		if p.Length() == 0 {
			break
		}
		if f := p.First(); len(strings.TrimSpace(f.Text())) == 0 {
			f.Remove()
		} else {
			break
		}
	}
	for {
		p := word.Find("p")
		if p.Length() == 0 {
			break
		}
		if f := p.Last(); len(strings.TrimSpace(f.Text())) == 0 {
			f.Remove()
		} else {
			break
		}
	}
	art.Content, _ = word.Html()
	art.Content = string(matchNote.ReplaceAll([]byte(art.Content), []byte{}))
	art.Content = strings.ReplaceAll(art.Content, "</p>", "</p>\n")
	art.Content = strings.TrimSpace(art.Content)
	if len(strings.TrimSpace(word.Text())) < 900 {
		return collect.ErrArticleTooShort
	}
	return nil
}

var ErrVideoArticle = errors.New("内容存在视频")
var matchNote = regexp.MustCompile(`(?U)<!--.+-->`)
