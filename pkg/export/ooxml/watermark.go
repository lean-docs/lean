package ooxml

import (
	"github.com/lean-docs/lean/pkg/ir"
	"strconv"
	"strings"
)

func writeWatermark(output *strings.Builder, mark *ir.Watermark, state *packageState) {
	decimal := func(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }
	style := "position:absolute;margin-left:" + decimal(mark.Left) + "pt;margin-top:" + decimal(mark.Top) + "pt;width:" + decimal(mark.Width) + "pt;height:" + decimal(mark.Height) + "pt;rotation:" + decimal(mark.Rotation) + ";z-index:-251654144;mso-position-horizontal:" + mark.HorizontalAlign + ";mso-position-horizontal-relative:" + mark.HorizontalRelative + ";mso-position-vertical:" + mark.VerticalAlign + ";mso-position-vertical-relative:" + mark.VerticalRelative
	if mark.Image != nil && state != nil {
		output.WriteString(`<w:pict xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office"><v:rect id="PictureWatermark` + strconv.Itoa(state.imageIDs[mark.Image]) + `" style="` + escape(style) + `" o:allowincell="f" fillcolor="#` + color(mark.Color) + `" stroked="f"><v:imagedata r:id="rIdImage` + strconv.Itoa(state.imageIDs[mark.Image]) + `" o:title="` + escape(mark.Image.Alt) + `"/><v:fill opacity="` + decimal(mark.Opacity) + `"/></v:rect></w:pict>`)
		return
	}
	output.WriteString(`<w:pict xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office"><v:shapetype id="_x0000_t136" coordsize="21600,21600" o:spt="136" path="m0,0l21600,0m0,21600l21600,21600e"><v:path textpathok="t"/><v:textpath on="t" fitshape="t"/></v:shapetype><v:shape id="PowerPlusWaterMarkObject1" type="#_x0000_t136" style="` + escape(style) + `" o:allowincell="f" fillcolor="#` + color(mark.Color) + `" stroked="f"><v:fill opacity="` + decimal(mark.Opacity) + `"/><v:textpath style="font-family:&quot;` + escape(mark.FontName) + `&quot;;font-size:1pt" string="` + escape(mark.Text) + `"/></v:shape></w:pict>`)
}
