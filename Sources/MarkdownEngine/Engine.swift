import Foundation
import Markdown

@_cdecl("render_markdown_to_html")
public func render_markdown_to_html(input: UnsafePointer<CChar>) -> UnsafeMutablePointer<CChar>? {
    let markdownString = String(cString: input)
    
    // Parse using swift-markdown
    let document = Document(parsing: markdownString)
    
    // Use the comprehensive visitor to walk the AST
    var visitor = HTMLVisitor()
    let htmlResult = visitor.visit(document)
    
    return strdup(htmlResult)
}

struct HTMLVisitor: MarkupVisitor {
    typealias Result = String
    
    func escapeHTML(_ string: String) -> String {
        return string
            .replacingOccurrences(of: "&", with: "&amp;")
            .replacingOccurrences(of: "<", with: "&lt;")
            .replacingOccurrences(of: ">", with: "&gt;")
            .replacingOccurrences(of: "\"", with: "&quot;")
            .replacingOccurrences(of: "'", with: "&#39;")
    }
    
    mutating func defaultVisit(_ markup: Markup) -> String {
        var result = ""
        for child in markup.children {
            result += visit(child)
        }
        return result
    }
    
    mutating func visitDocument(_ document: Document) -> String {
        return "<div class=\"markdown-body\">\(defaultVisit(document))</div>"
    }
    
    mutating func visitText(_ text: Text) -> String {
        return escapeHTML(text.string)
    }
    
    mutating func visitHeading(_ heading: Heading) -> String {
        let level = heading.level
        return "<h\(level)>\(defaultVisit(heading))</h\(level)>"
    }
    
    mutating func visitParagraph(_ paragraph: Paragraph) -> String {
        return "<p>\(defaultVisit(paragraph))</p>"
    }
    
    mutating func visitStrong(_ strong: Strong) -> String {
        return "<strong>\(defaultVisit(strong))</strong>"
    }
    
    mutating func visitEmphasis(_ emphasis: Emphasis) -> String {
        return "<em>\(defaultVisit(emphasis))</em>"
    }
    
    mutating func visitLink(_ link: Link) -> String {
        let dest = link.destination ?? ""
        let title = link.title ?? ""
        return "<a href=\"\(escapeHTML(dest))\" title=\"\(escapeHTML(title))\">\(defaultVisit(link))</a>"
    }
    
    mutating func visitOrderedList(_ orderedList: OrderedList) -> String {
        return "<ol>\(defaultVisit(orderedList))</ol>"
    }
    
    mutating func visitUnorderedList(_ unorderedList: UnorderedList) -> String {
        return "<ul>\(defaultVisit(unorderedList))</ul>"
    }
    
    mutating func visitListItem(_ listItem: ListItem) -> String {
        return "<li>\(defaultVisit(listItem))</li>"
    }
    
    mutating func visitCodeBlock(_ codeBlock: CodeBlock) -> String {
        let language = codeBlock.language ?? ""
        let langClass = language.isEmpty ? "" : " class=\"language-\(escapeHTML(language))\""
        return "<pre><code\(langClass)>\(escapeHTML(codeBlock.code))</code></pre>"
    }
    
    mutating func visitInlineCode(_ inlineCode: InlineCode) -> String {
        return "<code>\(escapeHTML(inlineCode.code))</code>"
    }
    
    mutating func visitBlockQuote(_ blockQuote: BlockQuote) -> String {
        return "<blockquote>\(defaultVisit(blockQuote))</blockquote>"
    }
    
    mutating func visitSoftBreak(_ softBreak: SoftBreak) -> String {
        return "\n"
    }
    
    mutating func visitLineBreak(_ lineBreak: LineBreak) -> String {
        return "<br>\n"
    }
    
    mutating func visitImage(_ image: Image) -> String {
        let src = image.source ?? ""
        var altText = ""
        // Recursively get plain text for alt attribute
        func extractPlainText(_ markup: Markup) -> String {
            if let text = markup as? Text {
                return text.string
            }
            return markup.children.map(extractPlainText).joined()
        }
        altText = image.children.map(extractPlainText).joined()
        
        let escapedSrc = escapeHTML(src)
        let escapedAlt = escapeHTML(altText)
        let title = image.title.map { " title=\"\(escapeHTML($0))\"" } ?? ""
        
        return "<img src=\"\(escapedSrc)\" alt=\"\(escapedAlt)\"\(title)>"
    }
    
    // Table Support
    mutating func visitTable(_ table: Table) -> String {
        return "<table>\(defaultVisit(table))</table>"
    }
    
    mutating func visitTableHead(_ tableHead: Table.Head) -> String {
        return "<thead>\(defaultVisit(tableHead))</thead>"
    }
    
    mutating func visitTableBody(_ tableBody: Table.Body) -> String {
        return "<tbody>\(defaultVisit(tableBody))</tbody>"
    }
    
    mutating func visitTableRow(_ tableRow: Table.Row) -> String {
        return "<tr>\(defaultVisit(tableRow))</tr>"
    }
    
    mutating func visitTableCell(_ tableCell: Table.Cell) -> String {
        let isHeader = tableCell.parent?.parent is Table.Head
        let tag = isHeader ? "th" : "td"
        return "<\(tag)>\(defaultVisit(tableCell))</\(tag)>"
    }
    
    mutating func visitThematicBreak(_ thematicBreak: ThematicBreak) -> String {
        return "<hr>"
    }

    mutating func visitHTMLBlock(_ html: HTMLBlock) -> String {
        return html.rawHTML
    }
    
    mutating func visitInlineHTML(_ inlineHTML: InlineHTML) -> String {
        return inlineHTML.rawHTML
    }
}
