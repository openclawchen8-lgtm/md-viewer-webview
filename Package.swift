// swift-tools-version:6.3
import PackageDescription

let package = Package(
    name: "MarkdownEngine",
    platforms: [.macOS(.v11)],
    products: [
        .library(name: "MarkdownEngine", type: .dynamic, targets: ["MarkdownEngine"]),
    ],
    dependencies: [
        .package(url: "https://github.com/apple/swift-markdown.git", branch: "main"),
    ],
    targets: [
        .target(
            name: "MarkdownEngine",
            dependencies: [
                .product(name: "Markdown", package: "swift-markdown"),
            ]),
    ]
)
