export function serializeXml(root: Element): string {
  return `<?xml version="1.0" encoding="UTF-8"?>\n${new XMLSerializer().serializeToString(root)}`
}

export function createXmlDocument(rootName: string, namespace: string): XMLDocument {
  return document.implementation.createDocument(namespace, rootName)
}
