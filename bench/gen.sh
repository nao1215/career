#!/bin/sh
# Writes the inputs of the himorime suite in bench/. Deterministic: the same
# arguments always write the same bytes, so a base and a head revision read
# the same input. Both revisions run the working tree's copy
# (${head_root}/gen.sh). All data is fictional.
#
#   sh gen.sh resume N FILE   a resume YAML with N employers: each is one
#                             history of three projects (the cv and the
#                             work-history) and a join and a leave line in
#                             work (the japanese-resume), every text in
#                             Japanese and English
set -eu

case "$1" in
resume)
	mkdir -p "$(dirname "$3")"
	awk -v n="$2" 'BEGIN {
		print "date: 2026年6月13日現在"
		print "theme:"
		print "  accent: \"#1f4e79\""
		print "profile:"
		print "  name: { ja: 見本 太郎, en: Taro Mihon }"
		print "  name_kana: みほん たろう"
		print "  birth_date: 1990年5月10日"
		print "  age: 満 36 歳"
		print "  gender: 男"
		print "  email: taro.mihon@example.com"
		print "  phone: \"+81 90-0000-0000\""
		print "  address:"
		print "    zip: 000-0000"
		print "    kana: とうきょうとみほんく"
		print "    text: { ja: 東京都見本区見本町1-2-3 見本ハイツ101, en: \"Tokyo, Japan\" }"
		print "education:"
		print "  - { year: 2009, month: 4, value: { ja: 見本工科大学 工学部 情報工学科 入学, en: Entered Mihon Institute of Technology } }"
		print "  - { year: 2013, month: 3, value: { ja: 見本工科大学 工学部 情報工学科 卒業, en: Graduated from Mihon Institute of Technology } }"
		print "work:"
		for (i = 1; i <= n; i++) {
			printf "  - { year: %d, month: 4, value: 見本商事第%d株式会社 入社 }\n", 2013 + i, i
			printf "  - { year: %d, month: 3, value: 見本商事第%d株式会社 退職 }\n", 2014 + i, i
		}
		print "  - { value: 現在に至る }"
		print "licenses:"
		print "  - { year: 2013, month: 6, value: 普通自動車第一種運転免許 取得 }"
		print "  - { year: 2015, month: 10, value: 応用情報技術者試験 取得 }"
		print "rireki:"
		print "  commuting_time: 60分"
		print "  dependents: ２人"
		print "  spouse: 有"
		print "  supporting_spouse: 有"
		print "  hobby: 個人アプリ開発と登山が趣味です。"
		print "  motivation: 貴社のプロダクトに強く共感し、志望しました。"
		print "  request: 貴社の規定に従います。"
		print "career:"
		print "  summary:"
		print "    ja: バックエンドとクラウドを中心に約10年の経験があります。"
		print "    en: About ten years across backend and cloud."
		print "  skills:"
		print "    - { ja: Go によるサーバーサイド開発, en: Backend development in Go }"
		print "    - { ja: クラウド基盤の設計と運用, en: Designing and running cloud platforms }"
		print "  histories:"
		for (i = n; i >= 1; i--) {
			printf "    - company: { ja: 見本商事第%d株式会社, en: Mihon Trading No.%d Inc. }\n", i, i
			printf "      period: { ja: %d年4月 - %d年3月, en: Apr %d - Mar %d }\n", 2013 + i, 2014 + i, 2013 + i, 2014 + i
			print "      role: { ja: ソフトウェアエンジニア, en: Software Engineer }"
			print "      summary:"
			print "        ja: 業務システムの設計から運用までを一貫して担当しました。"
			print "        en: Owned business systems from design through operation."
			print "      projects:"
			for (p = 1; p <= 3; p++) {
				printf "        - title: { ja: 業務システム第%d期の開発, en: Business system phase %d }\n", p, p
				printf "          period: { ja: %d年%d月 - %d年%d月, en: %d - %d }\n", 2013 + i, p * 3, 2013 + i, p * 3 + 2, 2013 + i, 2013 + i
				print "          role: { ja: 設計・実装（5名）, en: Design and implementation (team of 5) }"
				print "          description:"
				print "            ja: |"
				print "              受注から出荷までの業務を支えるシステムを設計・実装し、処理時間を"
				print "              半分に短縮しました。運用手順を整理し、障害対応の時間も減らしました。"
				print "            en: |"
				print "              Designed and built the system behind order to shipment, halving"
				print "              its processing time, and cut incident response time by tidying"
				print "              the runbooks."
				print "          tech: [Go, PostgreSQL, AWS]"
			}
		}
		print "  certifications:"
		print "    - { ja: 応用情報技術者試験, en: Applied Information Technology Engineer (Japan) }"
		print "  self_pr:"
		print "    ja: 設計から運用までを一人称で完遂することを得意としています。"
		print "    en: I own work from design through to operation."
	}' > "$3"
	;;
*)
	echo "gen.sh: unknown kind $1" >&2
	exit 2
	;;
esac
