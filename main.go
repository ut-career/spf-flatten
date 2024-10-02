package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/pkg/spflib"
)

const (
	TEMPLATE_HOST = "_spf.hitocolor.co.jp"
)

func main() {
	ctx := context.Background()

	// テンプレートを取得
	// この例では、_spf.example.comのTXTレコードから元になるSPFレコードを取得しているが、コマンドラインから入力を受け付けるようにしてもOK
	templateSPF, _ := FetchSPFRecordFromDomain(ctx, TEMPLATE_HOST)

	// 展開
	records, _ := GenerateSPFRecords(templateSPF)
	fmt.Println(records)
}

// GenerateSPFRecords SPFレコードを生成します
func GenerateSPFRecords(input string) (map[string]string, error) {
	// パース
	resolver := new(spflib.LiveResolver)
	r, err := spflib.Parse(input, resolver)
	if err != nil {
		return nil, err
	}

	// いくつかのドメインのみを展開する
	fr := r.Flatten(strings.Join([]string{
		// "spf.bmv.jp",
	}, ","))

	if err := removeDuplicated(fr); err != nil {
		return nil, err
	}

	// 長くなる場合を考慮しつつ展開
	//
	// 長い場合(255文字以上)は下記のように複数レコードに分割される
	// - @: "v=spf1 include:_spf1.example.com include:_spf2.example.com -all"
	// - _spf1: "v=spf1 ..... -all"
	// - _spf2: "v=spf1 ..... -all"
	rs := fr.TXTSplit("_spf%d.hitocolor.co.jp")
	return rs, nil
}

// removeDuplicated SPFレコードから重複したincludeを削除する
func removeDuplicated(r *spflib.SPFRecord) error {
	filtered := make([]*spflib.SPFPart, 0)
	remember := make(map[string]struct{})
	for _, v := range r.Parts {
		if v.IncludeDomain != "" {
			_, exist := remember[v.IncludeDomain]
			if exist {
				continue
			}
			remember[v.IncludeDomain] = struct{}{}
		}
		filtered = append(filtered, v)
	}
	r.Parts = filtered
	return nil
}

// FetchSPFRecordFromDomain ドメインのSPFレコードを取得します
func FetchSPFRecordFromDomain(ctx context.Context, domain string) (string, error) {
	r := new(spflib.LiveResolver)
	return r.GetSPF(domain)
}
