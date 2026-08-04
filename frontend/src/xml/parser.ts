export function parseXml(source: string): XMLDocument {
  const document = new DOMParser().parseFromString(source, 'application/xml')
  const parserError = document.querySelector('parsererror')

  if (parserError) {
    throw new Error('서버에서 올바르지 않은 XML 응답을 받았습니다.')
  }

  return document
}

export function textOf(parent: ParentNode, name: string): string {
  return parent.querySelector(name)?.textContent?.trim() ?? ''
}

export function booleanAttribute(element: Element, name: string): boolean {
  return element.getAttribute(name) === 'true'
}
