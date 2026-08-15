export function containsGmailAddress(value: string): boolean {
	return /@(?:gmail|googlemail)\.com(?=$|[\s,;>])/i.test(value)
}
