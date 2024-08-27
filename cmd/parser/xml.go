package main

import (
	"encoding/xml"
	"time"
)

// DiaDocXML - основное описание структуры XML файла
type DiaDocXML struct {
	XMLName  xml.Name          `xml:"Файл"`
	FileId   string            `xml:"ИдФайл,attr"`
	FormVer  string            `xml:"ВерсФорм,attr"`
	ProgVer  string            `xml:"ВерсПрог,attr"`
	Document DiaDocDocumentXML `xml:"Документ"`
}

// DiaDocDocumentXML - описание раздела Документ
type DiaDocDocumentXML struct {
	XMLName  xml.Name           `xml:"Документ"`
	Date     DiaDocDocumentDate `xml:"ДатаИнфПр,attr"`
	Seller   string             `xml:"НаимЭконСубСост,attr"`
	Products []DiaDocProductXML `xml:"ТаблСчФакт>СведТов"`
}

// DiaDocDocumentDate - описывает дату документа
type DiaDocDocumentDate struct {
	time.Time
}

// UnmarshalXMLAttr распознает дату и переводит её в тип Time
func (dd *DiaDocDocumentDate) UnmarshalXMLAttr(attr xml.Attr) error {
	const dateFormat = "02.01.2006"
	date, err := time.Parse(dateFormat, attr.Value)
	if err != nil {
		return err
	}
	*dd = DiaDocDocumentDate{date}
	return nil
}

// - описание позиции товара
type DiaDocProductXML struct {
	XMLName    xml.Name             `xml:"СведТов"`
	Name       string               `xml:"НаимТов,attr"`
	Count      int                  `xml:"КолТов,attr"`
	TotalPrice float32              `xml:"СтТовУчНал,attr"`
	ExtInfo    DiadocProductExtInfo `xml:"ДопСведТов"`
}

type DiadocProductExtInfo struct {
	XMLName xml.Name `xml:"ДопСведТов"`
	Code    string   `xml:"КодТов,attr"`
}
