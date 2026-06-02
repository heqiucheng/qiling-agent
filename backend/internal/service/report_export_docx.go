package service

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/heqiucheng/qiling-agent/backend/internal/domain"
)

const reportDOCXContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"

func renderReportDOCX(report domain.Report) ([]byte, error) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)

	files := map[string]string{
		"[Content_Types].xml":          docxContentTypesXML(),
		"_rels/.rels":                  docxRootRelsXML(),
		"word/_rels/document.xml.rels": docxDocumentRelsXML(),
		"word/styles.xml":              docxStylesXML(),
		"word/document.xml":            docxDocumentXML(report),
	}
	for path, content := range files {
		writer, err := archive.Create(path)
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write([]byte(content)); err != nil {
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func docxContentTypesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
  <Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>
</Types>`
}

func docxRootRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
}

func docxDocumentRelsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`
}

func docxStylesXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:style w:type="paragraph" w:default="1" w:styleId="Normal">
    <w:name w:val="Normal"/>
    <w:rPr><w:sz w:val="22"/><w:szCs w:val="22"/></w:rPr>
  </w:style>
  <w:style w:type="paragraph" w:styleId="Title">
    <w:name w:val="Title"/>
    <w:basedOn w:val="Normal"/>
    <w:rPr><w:b/><w:sz w:val="36"/><w:szCs w:val="36"/></w:rPr>
  </w:style>
  <w:style w:type="paragraph" w:styleId="Heading1">
    <w:name w:val="heading 1"/>
    <w:basedOn w:val="Normal"/>
    <w:rPr><w:b/><w:sz w:val="28"/><w:szCs w:val="28"/></w:rPr>
  </w:style>
</w:styles>`
}

func docxDocumentXML(report domain.Report) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	builder.WriteString(docxParagraph(report.Title, "Title"))
	builder.WriteString(docxParagraph(report.RangeLabel+" · "+report.GeneratedAt, ""))
	builder.WriteString(docxParagraph(report.Summary, ""))

	builder.WriteString(docxParagraph("指标概览", "Heading1"))
	for _, metric := range report.Metrics {
		builder.WriteString(docxParagraph(fmt.Sprintf("%s：%v。%s", metric.Label, metric.Value, metric.Hint), ""))
	}

	builder.WriteString(docxParagraph("客户分析", "Heading1"))
	for _, section := range report.Sections {
		builder.WriteString(docxParagraph(section.Title, "Heading1"))
		builder.WriteString(docxParagraph(section.Summary, ""))
		for _, item := range section.Items {
			builder.WriteString(docxParagraph(item.CustomerName+" ｜ "+item.Intent+" ｜ "+item.Stage, ""))
			builder.WriteString(docxParagraph("建议动作："+item.RecommendedAction, ""))
			builder.WriteString(docxParagraph("推荐话术："+item.Script, ""))
			builder.WriteString(docxParagraph("判断依据："+item.Reasoning, ""))
			if len(item.Evidence) > 0 {
				builder.WriteString(docxParagraph("证据："+strings.Join(item.Evidence, "；"), ""))
			}
		}
	}

	builder.WriteString(docxParagraph("行动清单", "Heading1"))
	if len(report.ActionItems) == 0 {
		builder.WriteString(docxParagraph("当前没有需要处理的行动项。", ""))
	} else {
		for _, item := range report.ActionItems {
			builder.WriteString(docxParagraph(fmt.Sprintf("[%s] %s：%s（%s）", item.Priority, item.CustomerName, item.Action, item.DueHint), ""))
		}
	}

	builder.WriteString(`<w:sectPr><w:pgSz w:w="11906" w:h="16838"/><w:pgMar w:top="1440" w:right="1440" w:bottom="1440" w:left="1440"/></w:sectPr>`)
	builder.WriteString(`</w:body></w:document>`)
	return builder.String()
}

func docxParagraph(text string, style string) string {
	var builder strings.Builder
	builder.WriteString("<w:p>")
	if style != "" {
		builder.WriteString(`<w:pPr><w:pStyle w:val="`)
		builder.WriteString(xmlEscape(style))
		builder.WriteString(`"/></w:pPr>`)
	}
	builder.WriteString("<w:r><w:t>")
	builder.WriteString(xmlEscape(text))
	builder.WriteString("</w:t></w:r></w:p>")
	return builder.String()
}

func xmlEscape(value string) string {
	var buffer bytes.Buffer
	_ = xml.EscapeText(&buffer, []byte(value))
	return buffer.String()
}
