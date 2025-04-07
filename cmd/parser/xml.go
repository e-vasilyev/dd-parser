package main

import (
	"encoding/xml"
	"time"
)

// DiaDocXML - основное описание структуры XML файла
type DiaDocXML struct {
	XMLName  xml.Name          `xml:"Файл"`
	FileID   string            `xml:"ИдФайл,attr"`
	FormVer  string            `xml:"ВерсФорм,attr"`
	ProgVer  string            `xml:"ВерсПрог,attr"`
	Document DiaDocDocumentXML `xml:"Документ"`
}

// DiaDocDocumentXML - описание раздела Документ
type DiaDocDocumentXML struct {
	XMLName  xml.Name                  `xml:"Документ"`
	Seller   string                    `xml:"НаимЭконСубСост,attr"`
	Invoice  DiadockDocumentInvoiceXML `xml:"СвСчФакт"`
	Products []DiaDocProductXML        `xml:"ТаблСчФакт>СведТов"`
}

// DiaDocDocumentDate - описывает дату документа
type DiaDocDocumentDate struct {
	time.Time
}

// DiadockDocumentInvoiceXML - описывает Счет фактуру
type DiadockDocumentInvoiceXML struct {
	XMLName     xml.Name `xml:"СвСчФакт"`
	Number      string
	NumberV5_01 string `xml:"НомерСчФ,attr"`
	NumberV5_03 string `xml:"НомерДок,attr"`
	Date        DiaDocDocumentDate
	DateV5_01   DiaDocDocumentDate `xml:"ДатаСчФ,attr"`
	DateV5_03   DiaDocDocumentDate `xml:"ДатаДок,attr"`
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

// DiaDocProductXML - описание позиции товара
type DiaDocProductXML struct {
	XMLName    xml.Name             `xml:"СведТов"`
	Name       string               `xml:"НаимТов,attr"`
	Count      int                  `xml:"КолТов,attr"`
	TotalPrice float32              `xml:"СтТовУчНал,attr"`
	ExtInfo    DiadocProductExtInfo `xml:"ДопСведТов"`
}

// DiadocProductExtInfo - описание дополнительный сведений по товару
type DiadocProductExtInfo struct {
	XMLName xml.Name `xml:"ДопСведТов"`
	Code    string   `xml:"КодТов,attr"`
}
