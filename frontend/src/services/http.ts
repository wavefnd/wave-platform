import { booleanAttribute, parseXml, textOf } from '../xml/parser'
import type {
  SourceBlob,
  SourceCommitDetail,
  SourceCommit,
  SourceEntry,
  SourceLanguage,
  SourceRefs,
  SourceRepository,
  SourceTree,
} from '../components/source/types'

export interface PlatformStatus {
  name: string
  status: string
  version: string
  environment: string
}

export interface ModuleStatus {
  name: string
  enabled: boolean
  status: string
}

export interface BlogPostSummary {
	slug: string
	category: 'article' | 'release' | 'roadmap'
	roadmapStatus: 'planned' | 'in-progress' | 'released' | ''
	roadmapOrder: number
	targetDate: string
	title: string
	summary: string
	status: 'draft' | 'published' | ''
	authorName: string
	publishedAt: string
	updatedAt: string
}

export interface BlogPost extends BlogPostSummary {
	content: string
	authorAccountId: string
	createdAt: string
}

export interface BlogPostInput {
	slug: string
	category: 'article' | 'release' | 'roadmap'
	roadmapStatus: 'planned' | 'in-progress' | 'released' | ''
	roadmapOrder: number
	targetDate: string
	title: string
	summary: string
	content: string
	status: 'draft' | 'published'
}

export type EditorCommand = 'bold' | 'italic' | 'inline-code' | 'heading' | 'quote' | 'unordered-list' | 'link'

export interface EditorTransformResult {
	content: string
	selectionStart: number
	selectionEnd: number
	engine: string
	lines: number
	words: number
}

export interface DocumentSummary {
  id: string
  path: string
  locale: DocumentLocale
  group: string
  order: number
  title: string
  summary: string
}

export type DocumentLocale = 'en' | 'ko' | 'ja' | 'zh' | 'es' | 'de' | 'ru' | 'id' | 'vi'

export interface DocumentBlock {
  kind: 'heading' | 'paragraph' | 'note' | 'warning' | 'code' | 'list' | 'table'
  anchor: string
  level: number
  language: string
  title: string
  text: string
  items: string[]
  rows: Array<{ header: boolean; cells: string[] }>
}

export interface DocumentView extends DocumentSummary {
  updatedAt: string
  markdown: string
  blocks: DocumentBlock[]
}

export interface PlatformStats {
  accounts: number
  messagesToday: number
  gitMirrors: number
}

export interface PatchSummary {
	id: string
	messageId: string
	subject: string
	title: string
	authorName: string
	authorEmail: string
	body: string
	preview: string
	version: number
	part: number
	total: number
	files: string[]
	sha256: string
	reviewStatus: 'received' | 'reviewing' | 'accepted' | 'rejected' | 'applied'
	targetRepository: string
	assigneeAccountId: string
	assigneeName: string
	reviewUpdatedAt: string
	seriesCount: number
	reviewCommentCount: number
	reviewComments: PatchReviewComment[]
	receivedAt: string
}

export interface PatchArchiveView { address: string; patches: PatchSummary[] }

export type RFCStatus = 'draft' | 'discussion' | 'accepted' | 'rejected' | 'implementing' | 'completed' | 'withdrawn'

export interface RFCComment {
	id: string
	authorAccountId: string
	authorName: string
	body: string
	createdAt: string
}

export interface RFCProposal {
	number: number
	title: string
	summary: string
	content: string
	status: RFCStatus
	authorAccountId: string
	authorName: string
	commentCount: number
	comments: RFCComment[]
	createdAt: string
	updatedAt: string
}

export interface PatchReviewComment {
	id: string
	authorAccountId: string
	authorName: string
	path: string
	line: number
	lineText: string
	body: string
	resolved: boolean
	createdAt: string
	updatedAt: string
}

export interface AdminAccount {
	id: string
	username: string
	displayName: string
	email: string
	status: string
	owner: boolean
	administrator: boolean
	sourceMaintainer: boolean
	rfcMaintainer: boolean
	totpEnabled: boolean
	recoveryVerified: boolean
	createdAt: string
	updatedAt: string
}

export interface AdminDelivery {
	id: string
	messageId: string
	recipient: string
	destination: string
	status: string
	attempts: number
	nextAttemptAt: string
	lastAttemptAt: string
	lastError: string
	createdAt: string
}

export interface AdminAuditEvent {
	id: string
	actorId: string
	resourceId: string
	action: string
	result: string
	occurredAt: string
}

export interface WebhookEndpoint {
	id: string
	scope: 'account' | 'platform' | ''
	name: string
	kind: 'generic' | 'discord'
	events: string[]
	destination: string
	enabled: boolean
	createdAt: string
	updatedAt: string
	signingSecret: string
}

export interface WebhookDelivery {
	id: string
	endpointId: string
	eventType: string
	title: string
	status: string
	attempts: number
	httpStatus: number
	lastError: string
	createdAt: string
	lastAttemptAt: string
}

export interface WebhookAdminView {
	supportedEvents: string[]
	endpoints: WebhookEndpoint[]
	deliveries: WebhookDelivery[]
}

export interface WebhookInput {
	id: string
	name: string
	kind: 'generic' | 'discord'
	url: string
	events: string[]
	enabled: boolean
	rotateSecret: boolean
}

export interface AdminSnapshot {
	accounts: AdminAccount[]
	security: {
		activeAccounts: number
		suspendedAccounts: number
		totpAccounts: number
		verifiedRecoveries: number
		registrationOpen: boolean
		turnstileEnabled: boolean
	}
	storage: { databaseBytes: number; valueLogBytes: number; filesBytes: number; health: string }
	mail: { queued: number; delivering: number; deferred: number; failed: number; delivered: number }
	deliveries: AdminDelivery[]
	gitMirrors: SourceRepository[]
	gitSyncInterval: string
	auditLog: AdminAuditEvent[]
	generatedAt: string
	lunaStevTimeZone: string
}

export interface AuthConfig {
	mailDomain: string
	registrationOpen: boolean
	totpConfigured: boolean
	turnstileSiteKey: string
}

export interface TOTPEnrollment {
	token: string
	secret: string
	uri: string
	expiresAt: string
}

export interface AccountSecurity {
	totpEnabled: boolean
	recoveryEmail: string
	recoveryVerified: boolean
}

export interface SponsorMember {
	name: string
	profile: string
	imageUrl: string
	website: string
	type: string
	amount: number
	currency: string
}

export interface SponsorTier {
	name: string
	slug: string
	amount: number
	currency: string
	interval: string
	members: SponsorMember[]
}

export interface SponsorsView {
	name: string
	url: string
	refreshedAt: string
	tiers: SponsorTier[]
}

export interface AccountSession {
	id: string
	username: string
	displayName: string
	email: string
	timeZone: string
	administrator: boolean
	owner: boolean
	sourceMaintainer: boolean
	rfcMaintainer: boolean
	expiresAt: string
}

export interface MailboxItem {
	id: string
	messageId: string
	from: string
	to: string[]
	subject: string
	receivedAt: string
	preview: string
	flags: string[]
	deliveryStatus: string
}

export interface MailboxView {
	address: string
	addresses: string[]
	folder: string
	items: MailboxItem[]
}

export interface MailingListSummary {
	id: string
	address: string
	name: string
	description: string
	postingPolicy: 'members' | 'staff'
	webhookPolicy: 'disabled' | 'summary' | 'full'
	webhookPreviewLimit: number
	subscribed: boolean
}

export interface MailingListThreadSummary {
	id: string
	listId: string
	rootMessageId: string
	subject: string
	preview: string
	author: string
	authorAccountId: string
	messageCount: number
	createdAt: string
	lastActivityAt: string
}

export interface MailingListMessage {
	id: string
	entryId: string
	messageId: string
	headerMessageId: string
	parentMessageId: string
	authorAccountId: string
	from: string
	to: string[]
	subject: string
	body: string
	createdAt: string
}

export interface MailingListThread {
	id: string
	listId: string
	address: string
	subject: string
	messages: MailingListMessage[]
}

export interface UserActivity {
	kind: 'community-post' | 'community-comment' | 'question' | 'answer'
	title: string
	excerpt: string
	url: string
	createdAt: string
}

export interface UserProfile {
	username: string
	displayName: string
	email: string
	bio: string
	timeZone: string
	joinedAt: string
	activities: UserActivity[]
	addressChoiceAllowed: boolean
}

export interface MailMessageView {
	entryId: string
	messageId: string
	from: string
	to: string[]
	cc: string[]
	subject: string
	date: string
	body: string
	flags: string[]
	deliveryStatus: string
}

export interface CommunitySpace {
  id: string
  slug: string
  name: string
  visibility: string
  postingPolicy: 'members' | 'owner'
}

export interface CommunityThreadSummary {
  id: string
  spaceId: string
  title: string
  author: string
	authorAccountId: string
  excerpt: string
  createdAt: string
  replyCount: number
	viewCount: number
	score: number
	viewerVote: number
	lastActivityAt: string
  tags: string[]
  pinned: boolean
  locked: boolean
}

export interface CommunityMessage {
  id: string
  parentMessageId: string
	authorAccountId: string
  author: string
  createdAt: string
  body: string
	score: number
	viewerVote: number
}

export interface CommunityThread {
  id: string
  spaceId: string
  title: string
  tags: string[]
  pinned: boolean
  locked: boolean
  root: CommunityMessage
  replies: CommunityMessage[]
	score: number
	viewCount: number
	viewerVote: number
	subscribed: boolean
}

export interface MediaUpload {
	id: string
	url: string
	width: number
	height: number
	bytes: number
}

export interface QuestionSummary {
	id: string
	title: string
	excerpt: string
	author: string
	authorAccountId: string
	createdAt: string
	lastActivityAt: string
	tags: string[]
	waveVersion: string
	platform: string
	status: string
	score: number
	viewerVote: number
	answerCount: number
	viewCount: number
	accepted: boolean
}

export interface QuestionMessage {
	id: string
	parentMessageId: string
	authorAccountId: string
	author: string
	createdAt: string
	body: string
	score: number
	viewerVote: number
	accepted: boolean
}

export interface QuestionView {
	id: string
	title: string
	status: string
	tags: string[]
	waveVersion: string
	platform: string
	acceptedMessageId: string
	root: QuestionMessage
	answers: QuestionMessage[]
	score: number
	viewCount: number
	viewerVote: number
}

async function getXml(path: string): Promise<XMLDocument> {
  const response = await fetch(path, {
    headers: { Accept: 'application/xml' },
  })

  const body = await response.text()
  if (!response.ok) {
    throw new Error(`요청 실패 (${response.status})`)
  }

  return parseXml(body)
}

async function requestXml(path: string, method: string, document?: XMLDocument): Promise<XMLDocument | null> {
	const response = await fetch(path, {
		method,
		credentials: 'same-origin',
		headers: {
			Accept: 'application/xml',
			...(document ? { 'Content-Type': 'application/xml; charset=utf-8' } : {}),
		},
		body: document ? new XMLSerializer().serializeToString(document) : undefined,
	})
	const body = await response.text()
	if (!response.ok) {
		if (body) {
			try {
				const xml = parseXml(body)
				throw new Error(textOf(xml, 'message') || `Request failed (${response.status})`)
			} catch (error) {
				if (error instanceof Error && !error.message.includes('올바르지 않은 XML')) throw error
			}
		}
		throw new Error(`Request failed (${response.status})`)
	}
	return body ? parseXml(body) : null
}

function authDocument(name: string, values: Record<string, string>): XMLDocument {
	const namespace = 'https://wave-lang.dev/ns/platform/api/v1'
	const document = window.document.implementation.createDocument(namespace, name)
	for (const [key, value] of Object.entries(values)) {
		const element = document.createElementNS(namespace, key)
		element.textContent = value
		document.documentElement.append(element)
	}
	return document
}

function parseAccountSession(xml: XMLDocument): AccountSession {
	return {
		id: textOf(xml, 'id'),
		username: textOf(xml, 'username'),
		displayName: textOf(xml, 'display-name'),
		email: textOf(xml, 'email'),
		timeZone: textOf(xml, 'time-zone') || 'UTC',
		administrator: textOf(xml, 'administrator') === 'true',
		owner: textOf(xml, 'owner') === 'true',
		sourceMaintainer: textOf(xml, 'source-maintainer') === 'true',
		rfcMaintainer: textOf(xml, 'rfc-maintainer') === 'true',
		expiresAt: textOf(xml, 'expires-at'),
	}
}

export async function getAuthConfig(): Promise<AuthConfig> {
	const xml = await getXml('/api/v1/auth/config')
	return {
		mailDomain: textOf(xml, 'mail-domain'),
		registrationOpen: textOf(xml, 'registration-open') === 'true',
		totpConfigured: textOf(xml, 'totp-configured') === 'true',
		turnstileSiteKey: textOf(xml, 'turnstile-site-key'),
	}
}

export async function getRegistrationAddress(displayName: string): Promise<{ localPart: string; choiceRequired: boolean }> {
	const query = new URLSearchParams({ 'display-name': displayName })
	const xml = await getXml(`/api/v1/auth/registration-address?${query}`)
	return { localPart: textOf(xml, 'local-part'), choiceRequired: textOf(xml, 'choice-required') === 'true' }
}

export async function getCurrentAccount(): Promise<AccountSession | null> {
	try {
		const xml = await getXml('/api/v1/auth/session')
		return parseAccountSession(xml)
	} catch (error) {
		if (error instanceof Error && error.message.includes('(401)')) return null
		throw error
	}
}

function parseEnrollment(xml: XMLDocument | null): TOTPEnrollment {
	if (!xml) throw new Error('The server returned an empty enrollment response.')
	return { token: textOf(xml, 'token'), secret: textOf(xml, 'secret'), uri: textOf(xml, 'uri'), expiresAt: textOf(xml, 'expires-at') }
}

export async function login(identifier: string, code: string, challenge = ''): Promise<AccountSession> {
	const xml = await requestXml('/api/v1/auth/login', 'POST', authDocument('login', { identifier, code, 'challenge-token': challenge }))
	if (!xml) throw new Error('The server returned an empty login response.')
	return parseAccountSession(xml)
}

export async function beginRegistration(displayName: string, username: string, recoveryEmail: string, challenge = ''): Promise<TOTPEnrollment> {
	return parseEnrollment(await requestXml('/api/v1/auth/register/begin', 'POST', authDocument('registration', {
		'display-name': displayName, username, 'recovery-email': recoveryEmail, 'challenge-token': challenge,
	})))
}

export async function finishRegistration(token: string, code: string): Promise<AccountSession> {
	const xml = await requestXml('/api/v1/auth/register/finish', 'POST', authDocument('enrollment', { token, code }))
	if (!xml) throw new Error('The server returned an empty registration response.')
	return parseAccountSession(xml)
}

export async function requestRecovery(identifier: string, challenge = ''): Promise<void> {
	await requestXml('/api/v1/auth/recovery/request', 'POST', authDocument('recovery', { identifier, 'challenge-token': challenge }))
}

export async function getRecoveryEnrollment(token: string): Promise<TOTPEnrollment> {
	return parseEnrollment(await requestXml('/api/v1/auth/recovery/enrollment', 'POST', authDocument('recovery', { token })))
}

export async function finishRecovery(token: string, code: string): Promise<void> {
	await requestXml('/api/v1/auth/recovery/finish', 'POST', authDocument('recovery', { token, code }))
}

export async function getAccountSecurity(): Promise<AccountSecurity> {
	const xml = await getXml('/api/v1/auth/security')
	return { totpEnabled: textOf(xml, 'totp-enabled') === 'true', recoveryEmail: textOf(xml, 'recovery-email'), recoveryVerified: textOf(xml, 'recovery-verified') === 'true' }
}

export async function beginTOTPRotation(code: string): Promise<TOTPEnrollment> {
	return parseEnrollment(await requestXml('/api/v1/auth/security/totp/begin', 'POST', authDocument('rotation', { code })))
}

export async function finishTOTPRotation(token: string, code: string): Promise<void> {
	await requestXml('/api/v1/auth/security/totp/finish', 'POST', authDocument('enrollment', { token, code }))
}

export async function changeRecoveryEmail(email: string, code: string): Promise<void> {
	await requestXml('/api/v1/auth/security/recovery-email', 'POST', authDocument('recovery-email', { email, code }))
}

export async function verifyRecoveryEmail(token: string): Promise<void> {
	await requestXml('/api/v1/auth/recovery-email/verify', 'POST', authDocument('verification', { token }))
}

export async function getSponsors(): Promise<SponsorsView> {
	const xml = await getXml('/api/v1/sponsors')
	return {
		name: textOf(xml, 'name'), url: textOf(xml, 'url'), refreshedAt: textOf(xml, 'refreshed-at'),
		tiers: Array.from(xml.querySelectorAll('tiers > tier')).map((tier) => ({
			name: childText(tier, 'name'), slug: childText(tier, 'slug'), amount: Number(childText(tier, 'amount')),
			currency: childText(tier, 'currency'), interval: childText(tier, 'interval'),
			members: Array.from(tier.querySelectorAll('members > member')).map((member) => ({
				name: childText(member, 'name'), profile: childText(member, 'profile'), imageUrl: childText(member, 'image-url'),
				website: childText(member, 'website'), type: childText(member, 'type'), amount: Number(childText(member, 'amount')),
				currency: childText(member, 'currency'),
			})),
		})),
	}
}

export async function logout(): Promise<void> {
	await requestXml('/api/v1/auth/logout', 'POST')
}

export async function getMailbox(folder = 'Inbox', q = ''): Promise<MailboxView> {
	const query = new URLSearchParams({ folder })
	if (q) query.set('q', q)
	const xml = await getXml(`/api/v1/mailbox?${query}`)
	return {
		address: textOf(xml, 'address'),
		addresses: Array.from(xml.querySelectorAll('addresses > address')).map((item) => item.textContent?.trim() ?? '').filter(Boolean),
		folder: textOf(xml, 'folder'),
		items: Array.from(xml.querySelectorAll('items > item')).map((item) => ({
			id: item.getAttribute('id') ?? '',
			messageId: childText(item, 'message-id'),
			from: childText(item, 'from'),
			to: Array.from(item.children).filter((child) => child.localName === 'to').map((child) => child.textContent?.trim() ?? '').filter(Boolean),
			subject: childText(item, 'subject'),
			receivedAt: childText(item, 'received-at'),
			preview: childContent(item, 'preview'),
			flags: Array.from(item.querySelectorAll('flags > flag')).map((flag) => flag.textContent?.trim() ?? '').filter(Boolean),
			deliveryStatus: childText(item, 'delivery-status'),
		})),
	}
}

function parseUserProfile(element: ParentNode): UserProfile {
	return {
		username: childText(element, 'username'), displayName: childText(element, 'display-name'),
		email: childText(element, 'email'), bio: childContent(element, 'bio'), timeZone: childText(element, 'time-zone') || 'UTC', joinedAt: childText(element, 'joined-at'),
		addressChoiceAllowed: childText(element, 'address-choice-allowed') === 'true',
		activities: Array.from(element.querySelectorAll('activities > activity')).map((item) => ({
			kind: childText(item, 'kind') as UserActivity['kind'], title: childText(item, 'title'),
			excerpt: childContent(item, 'excerpt'), url: childText(item, 'url'), createdAt: childText(item, 'created-at'),
		})),
	}
}

export async function getUsers(): Promise<UserProfile[]> {
	const xml = await getXml('/api/v1/users')
	return Array.from(xml.documentElement.children).filter((item) => item.localName === 'user' || item.localName === 'user-profile').map(parseUserProfile)
}

export async function getUser(username: string): Promise<UserProfile> {
	return parseUserProfile((await getXml(`/api/v1/users/${encodeURIComponent(username)}`)).documentElement)
}

export async function getUserByID(accountID: string): Promise<UserProfile> {
	return parseUserProfile((await getXml(`/api/v1/users/by-id/${encodeURIComponent(accountID)}`)).documentElement)
}

export async function updateUserProfile(displayName: string, bio: string, timeZone: string): Promise<UserProfile> {
	const xml = await requestXml('/api/v1/users/me/profile', 'POST', authDocument('user-profile-update', {
		'display-name': displayName, bio, 'time-zone': timeZone,
	}))
	if (!xml) throw new Error('The server returned an empty profile response.')
	return parseUserProfile(xml.documentElement)
}

export async function updateWaveAddress(localPart: string, code: string): Promise<UserProfile> {
	const xml = await requestXml('/api/v1/users/me/address', 'POST', authDocument('user-address-update', {
		'local-part': localPart, code,
	}))
	if (!xml) throw new Error('The server returned an empty profile response.')
	return parseUserProfile(xml.documentElement)
}

export async function getManagementMailbox(folder = 'Inbox', q = ''): Promise<MailboxView> {
	const query = new URLSearchParams({ folder })
	if (q) query.set('q', q)
	const xml = await getXml(`/api/v1/admin/mailbox?${query}`)
	return {
		address: textOf(xml, 'address'),
		addresses: Array.from(xml.querySelectorAll('addresses > address')).map((item) => item.textContent?.trim() ?? '').filter(Boolean),
		folder: textOf(xml, 'folder'),
		items: Array.from(xml.querySelectorAll('items > item')).map((item) => ({
			id: item.getAttribute('id') ?? '', messageId: childText(item, 'message-id'), from: childText(item, 'from'),
			to: Array.from(item.children).filter((child) => child.localName === 'to').map((child) => child.textContent?.trim() ?? '').filter(Boolean),
			subject: childText(item, 'subject'), receivedAt: childText(item, 'received-at'), preview: childContent(item, 'preview'),
			flags: Array.from(item.querySelectorAll('flags > flag')).map((flag) => flag.textContent?.trim() ?? '').filter(Boolean),
			deliveryStatus: childText(item, 'delivery-status'),
		})),
	}
}

export async function getManagementMailMessage(entryId: string): Promise<MailMessageView> {
	return parseMailMessage(await getXml(`/api/v1/admin/mailbox/messages/${encodeURIComponent(entryId)}`))
}

export async function updateManagementMailEntry(entryId: string, action: 'archive' | 'trash' | 'read' | 'unread'): Promise<MailMessageView> {
	const xml = await requestXml(`/api/v1/admin/mailbox/messages/${encodeURIComponent(entryId)}/action`, 'POST', authDocument('mailbox-action', { action }))
	if (!xml) throw new Error('The server returned an empty mailbox response.')
	return parseMailMessage(xml)
}

export async function sendManagementMail(from: string, to: string, subject: string, body: string, parentEntryId = ''): Promise<MailMessageView> {
	const xml = await requestXml('/api/v1/admin/mailbox/messages', 'POST', authDocument('send-management-mail', { from, to, subject, body, 'parent-entry-id': parentEntryId }))
	if (!xml) throw new Error('The server returned an empty team mail response.')
	return parseMailMessage(xml)
}

function parseMailMessage(xml: XMLDocument): MailMessageView {
	const root = xml.documentElement
	return {
		entryId: childText(root, 'entry-id'), messageId: childText(root, 'message-id'),
		from: childText(root, 'from'),
		to: Array.from(root.children).filter((child) => child.localName === 'to').map((child) => child.textContent?.trim() ?? '').filter(Boolean),
		cc: Array.from(root.children).filter((child) => child.localName === 'cc').map((child) => child.textContent?.trim() ?? '').filter(Boolean),
		subject: childText(root, 'subject'), date: childText(root, 'date'), body: childContent(root, 'body'),
		flags: Array.from(root.querySelectorAll('flags > flag')).map((flag) => flag.textContent?.trim() ?? '').filter(Boolean),
		deliveryStatus: childText(root, 'delivery-status'),
	}
}

export async function getMailMessage(entryId: string): Promise<MailMessageView> {
	return parseMailMessage(await getXml(`/api/v1/mailbox/messages/${encodeURIComponent(entryId)}`))
}

export async function sendMail(to: string, subject: string, body: string, parentEntryId = ''): Promise<MailMessageView> {
	const xml = await requestXml('/api/v1/mailbox/messages', 'POST', authDocument('send-mail', { to, subject, body, 'parent-entry-id': parentEntryId }))
	if (!xml) throw new Error('The server returned an empty mail response.')
	return parseMailMessage(xml)
}

export async function updateMailEntry(entryId: string, action: 'archive' | 'trash' | 'read' | 'unread'): Promise<MailMessageView> {
	const xml = await requestXml(`/api/v1/mailbox/messages/${encodeURIComponent(entryId)}/action`, 'POST', authDocument('mailbox-action', { action }))
	if (!xml) throw new Error('The server returned an empty mailbox response.')
	return parseMailMessage(xml)
}

function parseMailingList(element: Element): MailingListSummary {
	return {
		id: element.getAttribute('id') ?? '', address: childText(element, 'address'), name: childText(element, 'name'),
		description: childContent(element, 'description'), postingPolicy: childText(element, 'posting-policy') as MailingListSummary['postingPolicy'],
		webhookPolicy: childText(element, 'webhook-policy') as MailingListSummary['webhookPolicy'],
		webhookPreviewLimit: Number(childText(element, 'webhook-preview-limit')), subscribed: childText(element, 'subscribed') === 'true',
	}
}

function parseMailingListMessage(element: Element): MailingListMessage {
	return {
		id: element.getAttribute('id') ?? '', entryId: childText(element, 'entry-id'), messageId: childText(element, 'message-id'),
		headerMessageId: childText(element, 'header-message-id'), parentMessageId: childText(element, 'parent-message-id'),
		authorAccountId: childText(element, 'author-account-id'), from: childText(element, 'from'),
		to: Array.from(element.children).filter((child) => child.localName === 'to').map((child) => child.textContent?.trim() ?? '').filter(Boolean),
		subject: childText(element, 'subject'), body: childContent(element, 'body'), createdAt: childText(element, 'created-at'),
	}
}

function parseMailingListThread(xml: XMLDocument): MailingListThread {
	const root = xml.documentElement
	return {
		id: root.getAttribute('id') ?? '', listId: childText(root, 'list-id'), address: childText(root, 'address'),
		subject: childText(root, 'subject'),
		messages: Array.from(root.querySelectorAll('messages > message')).map(parseMailingListMessage),
	}
}

export async function getMailingLists(): Promise<MailingListSummary[]> {
	const xml = await getXml('/api/v1/mailing-lists')
	return Array.from(xml.documentElement.children).filter((item) => item.localName === 'list').map(parseMailingList)
}

export async function getMailingListThreads(list: string, q = '', limit = 30, offset = 0): Promise<MailingListThreadSummary[]> {
	const query = new URLSearchParams({ limit: String(limit), offset: String(offset) })
	if (q) query.set('q', q)
	const xml = await getXml(`/api/v1/mailing-lists/${encodeURIComponent(list)}/threads?${query}`)
	return Array.from(xml.documentElement.children).filter((item) => item.localName === 'thread').map((item) => ({
		id: item.getAttribute('id') ?? '', listId: childText(item, 'list-id'), rootMessageId: childText(item, 'root-message-id'),
		subject: childText(item, 'subject'), preview: childContent(item, 'preview'), author: childText(item, 'author'),
		authorAccountId: childText(item, 'author-account-id'), messageCount: Number(childText(item, 'message-count')),
		createdAt: childText(item, 'created-at'), lastActivityAt: childText(item, 'last-activity-at'),
	}))
}

export async function getMailingListThread(list: string, thread: string): Promise<MailingListThread> {
	return parseMailingListThread(await getXml(`/api/v1/mailing-lists/${encodeURIComponent(list)}/threads/${encodeURIComponent(thread)}`))
}

export async function setMailingListSubscription(list: string, subscribed: boolean): Promise<void> {
	await requestXml(`/api/v1/mailing-lists/${encodeURIComponent(list)}/subscription`, 'POST', authDocument('mailing-list-subscription', { subscribed: String(subscribed) }))
}

export async function postMailingListThread(list: string, subject: string, body: string): Promise<MailingListThread> {
	const xml = await requestXml(`/api/v1/mailing-lists/${encodeURIComponent(list)}/threads`, 'POST', authDocument('mailing-list-post', { subject, body }))
	if (!xml) throw new Error('The server returned an empty mailing list response.')
	return parseMailingListThread(xml)
}

export async function replyMailingListThread(list: string, thread: string, body: string, parentMessageId = ''): Promise<MailingListThread> {
	const xml = await requestXml(`/api/v1/mailing-lists/${encodeURIComponent(list)}/threads/${encodeURIComponent(thread)}/messages`, 'POST',
		authDocument('mailing-list-reply', { 'parent-message-id': parentMessageId, body }))
	if (!xml) throw new Error('The server returned an empty mailing list response.')
	return parseMailingListThread(xml)
}

const languageColors: Record<string, string> = {
  Wave: '#6654f1', Rust: '#dea584', Python: '#3572a5', Shell: '#89e051',
  Assembly: '#6e4c13', Makefile: '#427819', Dockerfile: '#384d54',
}

function childText(parent: ParentNode, name: string): string {
  return Array.from(parent.children).find((child) => child.localName === name)?.textContent?.trim() ?? ''
}

function childContent(parent: ParentNode, name: string): string {
  return Array.from(parent.children).find((child) => child.localName === name)?.textContent ?? ''
}

function parseCommit(element: Element | null): SourceCommit | undefined {
  if (!element) return undefined
  return {
    oid: childText(element, 'oid'),
    shortOid: childText(element, 'short-oid'),
    author: childText(element, 'author'),
    authoredAt: childText(element, 'authored-at'),
    subject: childText(element, 'subject'),
  }
}

function parseRepository(element: Element): SourceRepository {
  return {
    id: childText(element, 'id'), owner: childText(element, 'owner'),
    name: childText(element, 'name'), description: childText(element, 'description'),
    defaultBranch: childText(element, 'default-branch'), headOid: childText(element, 'head-oid'),
    status: childText(element, 'status'),
    headCommit: parseCommit(Array.from(element.children).find((child) => child.localName === 'head-commit') ?? null),
  }
}

function parseBlob(element: Element): SourceBlob {
  const highlight = Array.from(element.children).find((child) => child.localName === 'wave-highlight')
  return {
    path: childText(element, 'path'),
    oid: childText(element, 'oid'),
    size: Number(childText(element, 'size')),
    binary: childText(element, 'binary') === 'true',
    truncated: childText(element, 'truncated') === 'true',
    content: childContent(element, 'content'),
    waveHighlight: highlight ? {
      engine: highlight.getAttribute('engine') ?? 'wave',
      abi: Number(highlight.getAttribute('abi')) || 1,
      tokens: Array.from(highlight.getElementsByTagName('token')).map((token) => ({
        kind: token.getAttribute('kind') as 'keyword' | 'type' | 'string' | 'comment' | 'number',
        start: Number(token.getAttribute('start')),
        end: Number(token.getAttribute('end')),
      })),
    } : undefined,
  }
}

export async function getPlatformStatus(): Promise<PlatformStatus> {
  const xml = await getXml('/api/v1/platform')
  return {
    name: textOf(xml, 'name'),
    status: textOf(xml, 'status'),
    version: textOf(xml, 'version'),
    environment: textOf(xml, 'environment'),
  }
}

export async function getPlatformStats(): Promise<PlatformStats> {
  const xml = await getXml('/api/v1/platform/stats')
  return {
    accounts: Number(textOf(xml, 'accounts')) || 0,
    messagesToday: Number(textOf(xml, 'messages-today')) || 0,
    gitMirrors: Number(textOf(xml, 'git-mirrors')) || 0,
  }
}

function parsePatch(element: Element): PatchSummary {
	return {
		id: element.getAttribute('id') ?? '', messageId: childText(element, 'message-id'), subject: childText(element, 'subject'),
		title: childText(element, 'title'), authorName: childText(element, 'author-name'), authorEmail: childText(element, 'author-email'),
		body: childContent(element, 'body'), preview: childContent(element, 'preview'), version: Number(childText(element, 'version')) || 1,
		part: Number(childText(element, 'part')), total: Number(childText(element, 'total')),
		files: Array.from(element.querySelectorAll('files > file')).map((item) => item.textContent?.trim() ?? '').filter(Boolean),
		sha256: childText(element, 'sha256'), reviewStatus: (childText(element, 'review-status') || 'received') as PatchSummary['reviewStatus'],
		targetRepository: childText(element, 'target-repository'), assigneeAccountId: childText(element, 'assignee-account-id'),
		assigneeName: childText(element, 'assignee-name'), reviewUpdatedAt: childText(element, 'review-updated-at'),
		seriesCount: Number(childText(element, 'series-count')) || 1,
		reviewCommentCount: Number(childText(element, 'review-comment-count')) || 0,
		reviewComments: Array.from(element.querySelectorAll('review-comments > comment')).map(parsePatchReviewComment),
		receivedAt: childText(element, 'received-at'),
	}
}

function parsePatchReviewComment(element: Element): PatchReviewComment {
	return {
		id: element.getAttribute('id') ?? '', authorAccountId: childText(element, 'author-account-id'),
		authorName: childText(element, 'author-name'), path: childText(element, 'path'),
		line: Number(childText(element, 'line')), lineText: childContent(element, 'line-text'),
		body: childContent(element, 'body'), resolved: childText(element, 'resolved') === 'true',
		createdAt: childText(element, 'created-at'), updatedAt: childText(element, 'updated-at'),
	}
}

export async function getPatches(query = ''): Promise<PatchArchiveView> {
	const parameters = new URLSearchParams()
	if (query.trim()) parameters.set('q', query.trim())
	const xml = await getXml(`/api/v1/patches${parameters.size ? `?${parameters}` : ''}`)
	return { address: childText(xml.documentElement, 'address'), patches: Array.from(xml.documentElement.children).filter((item) => item.localName === 'patch').map(parsePatch) }
}

export async function getPatch(id: string): Promise<PatchSummary> {
	return parsePatch((await getXml(`/api/v1/patches/${encodeURIComponent(id)}`)).documentElement)
}

export async function updatePatchReview(id: string, status: PatchSummary['reviewStatus'], targetRepository: string): Promise<PatchSummary> {
	const xml = await requestXml(`/api/v1/patches/${encodeURIComponent(id)}/review`, 'POST', authDocument('patch-review', {
		status, 'target-repository': targetRepository,
	}))
	if (!xml) throw new Error('The server returned an empty patch review response.')
	return parsePatch(xml.documentElement)
}

export async function addPatchReviewComment(id: string, line: number, body: string): Promise<PatchReviewComment> {
	const xml = await requestXml(`/api/v1/patches/${encodeURIComponent(id)}/review-comments`, 'POST', authDocument('patch-review-comment', { line: String(line), body }))
	if (!xml) throw new Error('The server returned an empty patch review comment response.')
	return parsePatchReviewComment(xml.documentElement)
}

export async function resolvePatchReviewComment(id: string, commentId: string, resolved: boolean): Promise<PatchReviewComment> {
	const xml = await requestXml(`/api/v1/patches/${encodeURIComponent(id)}/review-comments/${encodeURIComponent(commentId)}/resolution`, 'POST', authDocument('patch-review-comment-resolution', { resolved: String(resolved) }))
	if (!xml) throw new Error('The server returned an empty patch review comment response.')
	return parsePatchReviewComment(xml.documentElement)
}

function parseRFCComment(element: Element): RFCComment {
	return {
		id: element.getAttribute('id') ?? '', authorAccountId: childText(element, 'author-account-id'),
		authorName: childText(element, 'author-name'), body: childContent(element, 'body'), createdAt: childText(element, 'created-at'),
	}
}

function parseRFC(element: Element): RFCProposal {
	return {
		number: Number(element.getAttribute('number')), title: childText(element, 'title'), summary: childContent(element, 'summary'),
		content: childContent(element, 'content'), status: (childText(element, 'status') || 'draft') as RFCStatus,
		authorAccountId: childText(element, 'author-account-id'), authorName: childText(element, 'author-name'),
		commentCount: Number(childText(element, 'comment-count')), comments: Array.from(element.querySelectorAll('comments > comment')).map(parseRFCComment),
		createdAt: childText(element, 'created-at'), updatedAt: childText(element, 'updated-at'),
	}
}

export async function getRFCs(query = '', status = ''): Promise<RFCProposal[]> {
	const parameters = new URLSearchParams()
	if (query.trim()) parameters.set('q', query.trim())
	if (status) parameters.set('status', status)
	const xml = await getXml(`/api/v1/rfcs${parameters.size ? `?${parameters}` : ''}`)
	return Array.from(xml.documentElement.children).filter((item) => item.localName === 'rfc').map(parseRFC)
}

export async function getRFC(number: number): Promise<RFCProposal> {
	return parseRFC((await getXml(`/api/v1/rfcs/${number}`)).documentElement)
}

export async function createRFC(title: string, content: string): Promise<RFCProposal> {
	const xml = await requestXml('/api/v1/rfcs', 'POST', authDocument('rfc', { title, content }))
	if (!xml) throw new Error('The server returned an empty RFC response.')
	return parseRFC(xml.documentElement)
}

export async function updateRFC(number: number, title: string, content: string): Promise<RFCProposal> {
	const xml = await requestXml(`/api/v1/rfcs/${number}`, 'POST', authDocument('rfc', { title, content }))
	if (!xml) throw new Error('The server returned an empty RFC response.')
	return parseRFC(xml.documentElement)
}

export async function updateRFCStatus(number: number, status: RFCStatus): Promise<RFCProposal> {
	const xml = await requestXml(`/api/v1/rfcs/${number}/status`, 'POST', authDocument('rfc-status', { status }))
	if (!xml) throw new Error('The server returned an empty RFC response.')
	return parseRFC(xml.documentElement)
}

export async function addRFCComment(number: number, body: string): Promise<RFCComment> {
	const xml = await requestXml(`/api/v1/rfcs/${number}/comments`, 'POST', authDocument('rfc-comment', { body }))
	if (!xml) throw new Error('The server returned an empty RFC comment response.')
	return parseRFCComment(xml.documentElement)
}

export async function getModules(): Promise<ModuleStatus[]> {
  const xml = await getXml('/api/v1/modules')
  return Array.from(xml.querySelectorAll('module')).map((element) => ({
    name: element.getAttribute('name') ?? '',
    enabled: booleanAttribute(element, 'enabled'),
    status: element.getAttribute('status') ?? 'unknown',
  }))
}

export async function getAdminSnapshot(): Promise<AdminSnapshot> {
	const xml = await getXml('/api/v1/admin')
	const root = xml.documentElement
	const security = Array.from(root.children).find((element) => element.localName === 'security')
	const storage = Array.from(root.children).find((element) => element.localName === 'storage')
	const mailStatus = Array.from(root.children).find((element) => element.localName === 'mail-status')
	return {
		accounts: Array.from(root.querySelectorAll('accounts > account')).map((element) => ({
			id: element.getAttribute('id') ?? '', username: childText(element, 'username'),
			displayName: childText(element, 'display-name'), email: childText(element, 'email'),
			status: childText(element, 'status'), owner: childText(element, 'owner') === 'true',
			administrator: childText(element, 'administrator') === 'true',
			sourceMaintainer: childText(element, 'source-maintainer') === 'true',
			rfcMaintainer: childText(element, 'rfc-maintainer') === 'true',
			totpEnabled: childText(element, 'totp-enabled') === 'true',
			recoveryVerified: childText(element, 'recovery-verified') === 'true',
			createdAt: childText(element, 'created-at'), updatedAt: childText(element, 'updated-at'),
		})),
		security: {
			activeAccounts: Number(security ? childText(security, 'active-accounts') : 0),
			suspendedAccounts: Number(security ? childText(security, 'suspended-accounts') : 0),
			totpAccounts: Number(security ? childText(security, 'totp-accounts') : 0),
			verifiedRecoveries: Number(security ? childText(security, 'verified-recoveries') : 0),
			registrationOpen: security ? childText(security, 'registration-open') === 'true' : false,
			turnstileEnabled: security ? childText(security, 'turnstile-enabled') === 'true' : false,
		},
		storage: {
			databaseBytes: Number(storage ? childText(storage, 'database-bytes') : 0),
			valueLogBytes: Number(storage ? childText(storage, 'value-log-bytes') : 0),
			filesBytes: Number(storage ? childText(storage, 'files-bytes') : 0),
			health: storage ? childText(storage, 'health') : 'unknown',
		},
		mail: {
			queued: Number(mailStatus ? childText(mailStatus, 'queued') : 0),
			delivering: Number(mailStatus ? childText(mailStatus, 'delivering') : 0),
			deferred: Number(mailStatus ? childText(mailStatus, 'deferred') : 0),
			failed: Number(mailStatus ? childText(mailStatus, 'failed') : 0),
			delivered: Number(mailStatus ? childText(mailStatus, 'delivered') : 0),
		},
		deliveries: Array.from(root.querySelectorAll('deliveries > delivery')).map((element) => ({
			id: element.getAttribute('id') ?? '', messageId: childText(element, 'message-id'),
			recipient: childText(element, 'recipient'), destination: childText(element, 'destination'),
			status: childText(element, 'status'), attempts: Number(childText(element, 'attempts')),
			nextAttemptAt: childText(element, 'next-attempt-at'), lastAttemptAt: childText(element, 'last-attempt-at'),
			lastError: childText(element, 'last-error'), createdAt: childText(element, 'created-at'),
		})),
		gitMirrors: Array.from(root.querySelectorAll('git-mirrors > repository')).map(parseRepository),
		gitSyncInterval: childText(root, 'git-sync-interval'),
		auditLog: Array.from(root.querySelectorAll('audit-log > event')).map((element) => ({
			id: element.getAttribute('id') ?? '', actorId: childText(element, 'actor-id'),
			resourceId: childText(element, 'resource-id'), action: childText(element, 'action'),
			result: childText(element, 'result'), occurredAt: childText(element, 'occurred-at'),
		})),
		generatedAt: childText(root, 'generated-at'),
		lunaStevTimeZone: childText(root, 'lunastev-time-zone') || 'Asia/Seoul',
	}
}

export async function getPlatformPreferences(): Promise<{ lunaStevTimeZone: string }> {
	const xml = await getXml('/api/v1/platform/preferences')
	return { lunaStevTimeZone: textOf(xml, 'lunastev-time-zone') || 'Asia/Seoul' }
}

export async function updateLunaStevTimeZone(timeZone: string): Promise<void> {
	await requestXml('/api/v1/admin/settings/lunastev-time-zone', 'POST', authDocument('time-zone-setting', { 'time-zone': timeZone }))
}

export async function updateAdminAccountStatus(accountId: string, status: 'active' | 'suspended'): Promise<void> {
	await requestXml(`/api/v1/admin/accounts/${encodeURIComponent(accountId)}/status`, 'POST', authDocument('account-status', { status }))
}

export async function updateAdminAccountRole(accountId: string, administrator: boolean): Promise<void> {
	await requestXml(`/api/v1/admin/accounts/${encodeURIComponent(accountId)}/role`, 'POST', authDocument('account-role', { administrator: String(administrator) }))
}

export async function updateAdminSourceMaintainer(accountId: string, enabled: boolean): Promise<void> {
	await requestXml(`/api/v1/admin/accounts/${encodeURIComponent(accountId)}/source-maintainer`, 'POST', authDocument('source-maintainer', { enabled: String(enabled) }))
}

export async function updateAdminRFCMaintainer(accountId: string, enabled: boolean): Promise<void> {
	await requestXml(`/api/v1/admin/accounts/${encodeURIComponent(accountId)}/rfc-maintainer`, 'POST', authDocument('rfc-maintainer', { enabled: String(enabled) }))
}

function parseWebhook(element: Element): WebhookEndpoint {
	return {
		id: element.getAttribute('id') ?? '', name: childText(element, 'name'),
		scope: childText(element, 'scope') as WebhookEndpoint['scope'],
		kind: childText(element, 'kind') === 'discord' ? 'discord' : 'generic',
		events: Array.from(element.querySelectorAll('events > event')).map((item) => item.textContent?.trim() ?? '').filter(Boolean),
		destination: childText(element, 'destination'), enabled: childText(element, 'enabled') === 'true',
		createdAt: childText(element, 'created-at'), updatedAt: childText(element, 'updated-at'),
		signingSecret: childText(element, 'signing-secret'),
	}
}

export async function getAdminWebhooks(): Promise<WebhookAdminView> {
	return parseWebhookAdminView(await getXml('/api/v1/admin/webhooks'))
}

function parseWebhookAdminView(xml: XMLDocument): WebhookAdminView {
	const root = xml.documentElement
	return {
		supportedEvents: Array.from(root.querySelectorAll('supported-events > event')).map((item) => item.textContent?.trim() ?? '').filter(Boolean),
		endpoints: Array.from(root.querySelectorAll('endpoints > webhook')).map(parseWebhook),
		deliveries: Array.from(root.querySelectorAll('deliveries > delivery')).map((element) => ({
			id: element.getAttribute('id') ?? '', endpointId: childText(element, 'endpoint-id'), eventType: childText(element, 'event-type'),
			title: childText(element, 'title'), status: childText(element, 'status'), attempts: Number(childText(element, 'attempts')),
			httpStatus: Number(childText(element, 'http-status')), lastError: childText(element, 'last-error'),
			createdAt: childText(element, 'created-at'), lastAttemptAt: childText(element, 'last-attempt-at'),
		})),
	}
}

export async function getAccountWebhooks(): Promise<WebhookAdminView> {
	return parseWebhookAdminView(await getXml('/api/v1/webhooks'))
}

export async function saveAdminWebhook(input: WebhookInput): Promise<WebhookEndpoint> {
	return saveWebhookAt('/api/v1/admin/webhooks', input)
}

async function saveWebhookAt(path: string, input: WebhookInput): Promise<WebhookEndpoint> {
	const document = authDocument('webhook', {
		id: input.id, name: input.name, kind: input.kind, url: input.url,
		enabled: String(input.enabled), 'rotate-secret': String(input.rotateSecret),
	})
	const namespace = document.documentElement.namespaceURI
	const events = document.createElementNS(namespace, 'events')
	for (const value of input.events) { const item = document.createElementNS(namespace, 'event'); item.textContent = value; events.append(item) }
	document.documentElement.append(events)
	const xml = await requestXml(path, 'POST', document)
	if (!xml) throw new Error('The server returned an empty webhook response.')
	return parseWebhook(xml.documentElement)
}

export async function saveAccountWebhook(input: WebhookInput): Promise<WebhookEndpoint> {
	return saveWebhookAt('/api/v1/webhooks', input)
}

export async function deleteAdminWebhook(id: string): Promise<void> {
	await requestXml(`/api/v1/admin/webhooks/${encodeURIComponent(id)}`, 'DELETE')
}

export async function testAdminWebhook(id: string): Promise<void> {
	await requestXml(`/api/v1/admin/webhooks/${encodeURIComponent(id)}/test`, 'POST', authDocument('webhook-test', {}))
}

export async function deleteAccountWebhook(id: string): Promise<void> {
	await requestXml(`/api/v1/webhooks/${encodeURIComponent(id)}`, 'DELETE')
}

export async function testAccountWebhook(id: string): Promise<void> {
	await requestXml(`/api/v1/webhooks/${encodeURIComponent(id)}/test`, 'POST', authDocument('webhook-test', {}))
}

function parseBlogSummary(element: ParentNode): BlogPostSummary {
	return {
		slug: childText(element, 'slug'),
		category: ['release', 'roadmap'].includes(childText(element, 'category')) ? childText(element, 'category') as BlogPostSummary['category'] : 'article',
		roadmapStatus: childText(element, 'roadmap-status') as BlogPostSummary['roadmapStatus'],
		roadmapOrder: Number(childText(element, 'roadmap-order')) || 0,
		targetDate: childText(element, 'target-release-date'),
		title: childText(element, 'title'), summary: childContent(element, 'summary'),
		status: childText(element, 'status') as BlogPostSummary['status'], authorName: childText(element, 'author-name'),
		publishedAt: childText(element, 'published-at'), updatedAt: childText(element, 'updated-at'),
	}
}

function parseBlogPost(element: ParentNode): BlogPost {
	return {
		...parseBlogSummary(element), content: childContent(element, 'content'),
		authorAccountId: childText(element, 'author-account-id'), createdAt: childText(element, 'created-at'),
	}
}

export async function getBlogPosts(category?: 'article' | 'release' | 'roadmap', limit = 0): Promise<BlogPostSummary[]> {
	const query = new URLSearchParams()
	if (category) query.set('category', category)
	if (limit > 0) query.set('limit', String(limit))
	const xml = await getXml(`/api/v1/blog/posts${query.size ? `?${query}` : ''}`)
	return Array.from(xml.documentElement.children).filter((item) => item.localName === 'post').map(parseBlogSummary)
}

export async function getBlogPost(slug: string): Promise<BlogPost> {
	return parseBlogPost((await getXml(`/api/v1/blog/posts/${encodeURIComponent(slug)}`)).documentElement)
}

export async function getBlogEditorPosts(): Promise<BlogPostSummary[]> {
	const xml = await getXml('/api/v1/blog/editor/posts')
	return Array.from(xml.documentElement.children).filter((item) => item.localName === 'post').map(parseBlogSummary)
}

export async function getBlogEditorPost(slug: string): Promise<BlogPost> {
	return parseBlogPost((await getXml(`/api/v1/blog/editor/posts/${encodeURIComponent(slug)}`)).documentElement)
}

export async function saveBlogEditorPost(input: BlogPostInput): Promise<BlogPost> {
	const xml = await requestXml('/api/v1/blog/editor/posts', 'POST', authDocument('blog-post', {
		slug: input.slug, category: input.category, 'roadmap-status': input.roadmapStatus,
		'roadmap-order': String(input.roadmapOrder),
		'target-release-date': input.targetDate,
		title: input.title, summary: input.summary, content: input.content, status: input.status,
	}))
	if (!xml) throw new Error('The server returned an empty blog response.')
	return parseBlogPost(xml.documentElement)
}

function unicodeOffset(value: string, browserOffset: number): number {
	return Array.from(value.slice(0, browserOffset)).length
}

function browserOffset(value: string, unicodePosition: number): number {
	return Array.from(value).slice(0, unicodePosition).join('').length
}

export async function transformEditorDocument(content: string, selectionStart: number, selectionEnd: number, command: EditorCommand): Promise<EditorTransformResult> {
	const xml = await requestXml('/api/v1/editor/transform', 'POST', authDocument('editor-transform', {
		content,
		'selection-start': String(unicodeOffset(content, selectionStart)),
		'selection-end': String(unicodeOffset(content, selectionEnd)),
		command,
	}))
	if (!xml) throw new Error('WaveEditor returned an empty response.')
	const root = xml.documentElement
	const transformed = childContent(root, 'content')
	return {
		content: transformed,
		selectionStart: browserOffset(transformed, Number(childText(root, 'selection-start')) || 0),
		selectionEnd: browserOffset(transformed, Number(childText(root, 'selection-end')) || 0),
		engine: childText(root, 'engine'),
		lines: Number(childText(root, 'lines')) || 1,
		words: Number(childText(root, 'words')) || 0,
	}
}

function parseDocumentSummary(element: Element): DocumentSummary {
  return {
    id: childText(element, 'id'),
    path: childText(element, 'path'),
    locale: childText(element, 'locale') as DocumentLocale,
    group: childText(element, 'group'),
    order: Number(childText(element, 'order')) || 0,
    title: childText(element, 'title'),
    summary: childText(element, 'summary'),
  }
}

export async function getDocuments(locale: DocumentLocale): Promise<DocumentSummary[]> {
  const xml = await getXml(`/api/v1/documents?locale=${locale}`)
  return Array.from(xml.documentElement.children)
    .filter((element) => element.localName === 'document')
    .map(parseDocumentSummary)
}

export async function getDocument(path: string, locale: DocumentLocale): Promise<DocumentView> {
  const xml = await getXml(`/api/v1/documents/${path.split('/').map(encodeURIComponent).join('/')}?locale=${locale}`)
  const root = xml.documentElement
  const content = Array.from(root.children).find((element) => element.localName === 'content')
  const blocks = content ? Array.from(content.children).filter((element) => element.localName === 'block').map((element) => ({
    kind: (element.getAttribute('kind') ?? 'paragraph') as DocumentBlock['kind'],
    anchor: element.getAttribute('anchor') ?? '',
    level: Number(element.getAttribute('level')) || 0,
    language: element.getAttribute('language') ?? '',
    title: element.getAttribute('title') ?? '',
    text: Array.from(element.childNodes).filter((node) => node.nodeType === Node.TEXT_NODE || node.nodeType === Node.CDATA_SECTION_NODE).map((node) => node.textContent ?? '').join('').trim(),
    items: Array.from(element.children).filter((child) => child.localName === 'item').map((child) => child.textContent?.trim() ?? ''),
    rows: Array.from(element.children).filter((child) => child.localName === 'row').map((row) => ({
      header: row.getAttribute('header') === 'true',
      cells: Array.from(row.children).filter((cell) => cell.localName === 'cell').map((cell) => cell.textContent?.trim() ?? ''),
    })),
  })) : []
  return {
    ...parseDocumentSummary(root),
    updatedAt: childText(root, 'updated-at'),
    markdown: content ? childText(content, 'markdown') : '',
    blocks,
  }
}

export async function getCommunitySpaces(): Promise<CommunitySpace[]> {
  const xml = await getXml('/api/v1/community/spaces')
  return Array.from(xml.documentElement.children)
    .filter((element) => element.localName === 'space')
    .map((element) => ({
      id: element.getAttribute('id') ?? '',
      slug: childText(element, 'slug'),
      name: childText(element, 'name'),
      visibility: childText(element, 'visibility'),
      postingPolicy: childText(element, 'posting-policy') === 'owner' ? 'owner' : 'members',
    }))
}

export async function getCommunityThreads(space = '', options: { sort?: string; q?: string; limit?: number; offset?: number } = {}): Promise<CommunityThreadSummary[]> {
  const query = new URLSearchParams()
	if (space) query.set('space', space)
	if (options.sort) query.set('sort', options.sort)
	if (options.q) query.set('q', options.q)
	if (options.limit) query.set('limit', String(options.limit))
	if (options.offset) query.set('offset', String(options.offset))
	const suffix = query.size ? `?${query}` : ''
  const xml = await getXml(`/api/v1/community/threads${suffix}`)
  return Array.from(xml.documentElement.children)
    .filter((element) => element.localName === 'thread')
    .map((element) => ({
      id: childText(element, 'id'),
      spaceId: childText(element, 'space-id'),
      title: childText(element, 'title'),
      author: childText(element, 'author'),
		authorAccountId: childText(element, 'author-account-id'),
      excerpt: childContent(element, 'excerpt'),
      createdAt: childText(element, 'created-at'),
      replyCount: Number(childText(element, 'reply-count')) || 0,
		viewCount: Number(childText(element, 'view-count')) || 0,
		score: Number(childText(element, 'score')) || 0,
		viewerVote: Number(childText(element, 'viewer-vote')) || 0,
		lastActivityAt: childText(element, 'last-activity-at'),
      tags: Array.from(element.querySelectorAll('tags > tag')).map((tag) => tag.textContent?.trim() ?? '').filter(Boolean),
      pinned: childText(element, 'pinned') === 'true',
      locked: childText(element, 'locked') === 'true',
    }))
}

function parseCommunityMessage(element: Element | undefined): CommunityMessage {
  return {
    id: element ? childText(element, 'id') : '',
    parentMessageId: element ? childText(element, 'parent-message-id') : '',
		authorAccountId: element ? childText(element, 'author-account-id') : '',
    author: element ? childText(element, 'author') : '',
    createdAt: element ? childText(element, 'created-at') : '',
    body: element ? childContent(element, 'body') : '',
		score: element ? Number(childText(element, 'score')) || 0 : 0,
		viewerVote: element ? Number(childText(element, 'viewer-vote')) || 0 : 0,
  }
}

export async function getCommunityThread(id: string): Promise<CommunityThread> {
  const xml = await getXml(`/api/v1/community/threads/${encodeURIComponent(id)}`)
	return parseCommunityThread(xml)
}

function communityDocument(name: string): XMLDocument {
	return window.document.implementation.createDocument('https://wave-lang.dev/ns/platform/api/v1', name)
}

function appendXmlValue(document: XMLDocument, parent: Element, name: string, value: string) {
	const element = document.createElementNS(document.documentElement.namespaceURI, name)
	element.textContent = value
	parent.append(element)
}

export async function createCommunityPost(input: { spaceId: string; title: string; body: string; tags: string[] }): Promise<CommunityThread> {
	const document = communityDocument('community-post')
	appendXmlValue(document, document.documentElement, 'space-id', input.spaceId)
	appendXmlValue(document, document.documentElement, 'title', input.title)
	appendXmlValue(document, document.documentElement, 'body', input.body)
	if (input.tags.length) {
		const tags = document.createElementNS(document.documentElement.namespaceURI, 'tags')
		for (const tag of input.tags) appendXmlValue(document, tags, 'tag', tag)
		document.documentElement.append(tags)
	}
	const xml = await requestXml('/api/v1/community/threads', 'POST', document)
	if (!xml) throw new Error('The server returned an empty post response.')
	return parseCommunityThread(xml)
}

export async function uploadLunaStevImage(file: File): Promise<MediaUpload> {
	const form = new FormData()
	form.append('image', file, file.name)
	const response = await fetch('/api/v1/media/lunastev/images', {
		method: 'POST', credentials: 'same-origin', headers: { Accept: 'application/xml' }, body: form,
	})
	const body = await response.text()
	if (!response.ok) {
		if (body) {
			try {
				const xml = parseXml(body)
				throw new Error(textOf(xml, 'message') || `Request failed (${response.status})`)
			} catch (error) {
				if (error instanceof Error && !error.message.includes('올바르지 않은 XML')) throw error
			}
		}
		throw new Error(`Request failed (${response.status})`)
	}
	const xml = parseXml(body)
	return {
		id: textOf(xml, 'id'), url: textOf(xml, 'url'),
		width: Number(textOf(xml, 'width')) || 0, height: Number(textOf(xml, 'height')) || 0,
		bytes: Number(textOf(xml, 'bytes')) || 0,
	}
}

export async function createCommunityReply(threadId: string, body: string, parentMessageId = ''): Promise<CommunityThread> {
	const document = communityDocument('community-reply')
	if (parentMessageId) appendXmlValue(document, document.documentElement, 'parent-message-id', parentMessageId)
	appendXmlValue(document, document.documentElement, 'body', body)
	const xml = await requestXml(`/api/v1/community/threads/${encodeURIComponent(threadId)}/comments`, 'POST', document)
	if (!xml) throw new Error('The server returned an empty reply response.')
	return parseCommunityThread(xml)
}

export async function voteCommunity(threadId: string, targetType: 'thread' | 'message', targetId: string, value: -1 | 0 | 1): Promise<{ score: number; viewerVote: number }> {
	const document = communityDocument('community-vote')
	appendXmlValue(document, document.documentElement, 'target-type', targetType)
	appendXmlValue(document, document.documentElement, 'target-id', targetId)
	appendXmlValue(document, document.documentElement, 'value', String(value))
	const xml = await requestXml(`/api/v1/community/threads/${encodeURIComponent(threadId)}/vote`, 'POST', document)
	if (!xml) throw new Error('The server returned an empty vote response.')
	return { score: Number(textOf(xml, 'score')) || 0, viewerVote: Number(textOf(xml, 'viewer-vote')) || 0 }
}

export async function subscribeCommunity(threadId: string, subscribed: boolean): Promise<void> {
	const document = communityDocument('community-subscription')
	appendXmlValue(document, document.documentElement, 'subscribed', String(subscribed))
	await requestXml(`/api/v1/community/threads/${encodeURIComponent(threadId)}/subscription`, 'POST', document)
}

function parseQuestionMessage(element: Element | undefined): QuestionMessage {
	return {
		id: element ? childText(element, 'id') : '',
		parentMessageId: element ? childText(element, 'parent-message-id') : '',
		authorAccountId: element ? childText(element, 'author-account-id') : '',
		author: element ? childText(element, 'author') : '',
		createdAt: element ? childText(element, 'created-at') : '',
		body: element ? childContent(element, 'body') : '',
		score: element ? Number(childText(element, 'score')) || 0 : 0,
		viewerVote: element ? Number(childText(element, 'viewer-vote')) || 0 : 0,
		accepted: element ? childText(element, 'accepted') === 'true' : false,
	}
}

function parseQuestionView(xml: XMLDocument): QuestionView {
	const root = xml.documentElement
	const metadata = Array.from(root.children).find((element) => element.localName === 'question')
	const rootMessage = Array.from(root.children).find((element) => element.localName === 'root')
	const answers = Array.from(root.querySelector('answers')?.children ?? [])
		.filter((element) => element.localName === 'answer').map((element) => parseQuestionMessage(element))
	return {
		id: metadata?.getAttribute('id') ?? '', title: childText(root, 'title'),
		status: metadata ? childText(metadata, 'status') : '',
		tags: Array.from(metadata?.querySelectorAll('tags > tag') ?? []).map((tag) => tag.textContent?.trim() ?? '').filter(Boolean),
		waveVersion: metadata ? childText(metadata, 'wave-version') : '',
		platform: metadata ? childText(metadata, 'platform') : '',
		acceptedMessageId: metadata ? childText(metadata, 'accepted-message-id') : '',
		root: parseQuestionMessage(rootMessage), answers,
		score: Number(childText(root, 'score')) || 0,
		viewCount: Number(childText(root, 'view-count')) || 0,
		viewerVote: Number(childText(root, 'viewer-vote')) || 0,
	}
}

export async function getQuestions(options: { sort?: string; q?: string; tag?: string; limit?: number; offset?: number } = {}): Promise<QuestionSummary[]> {
	const query = new URLSearchParams()
	if (options.sort) query.set('sort', options.sort)
	if (options.q) query.set('q', options.q)
	if (options.tag) query.set('tag', options.tag)
	if (options.limit) query.set('limit', String(options.limit))
	if (options.offset) query.set('offset', String(options.offset))
	const xml = await getXml(`/api/v1/questions${query.size ? `?${query}` : ''}`)
	return Array.from(xml.documentElement.children).filter((element) => element.localName === 'question').map((element) => ({
		id: childText(element, 'id'), title: childText(element, 'title'), excerpt: childContent(element, 'excerpt'),
		author: childText(element, 'author'), authorAccountId: childText(element, 'author-account-id'),
		createdAt: childText(element, 'created-at'), lastActivityAt: childText(element, 'last-activity-at'),
		tags: Array.from(element.querySelectorAll('tags > tag')).map((tag) => tag.textContent?.trim() ?? '').filter(Boolean),
		waveVersion: childText(element, 'wave-version'), platform: childText(element, 'platform'),
		status: childText(element, 'status'), score: Number(childText(element, 'score')) || 0,
		viewerVote: Number(childText(element, 'viewer-vote')) || 0,
		answerCount: Number(childText(element, 'answer-count')) || 0,
		viewCount: Number(childText(element, 'view-count')) || 0,
		accepted: childText(element, 'accepted') === 'true',
	}))
}

export async function getQuestion(id: string): Promise<QuestionView> {
	return parseQuestionView(await getXml(`/api/v1/questions/${encodeURIComponent(id)}`))
}

export async function createQuestion(input: { title: string; body: string; tags: string[]; waveVersion: string; platform: string }): Promise<QuestionView> {
	const document = communityDocument('question-create')
	appendXmlValue(document, document.documentElement, 'title', input.title)
	appendXmlValue(document, document.documentElement, 'body', input.body)
	const tags = document.createElementNS(document.documentElement.namespaceURI, 'tags')
	for (const tag of input.tags) appendXmlValue(document, tags, 'tag', tag)
	document.documentElement.append(tags)
	if (input.waveVersion) appendXmlValue(document, document.documentElement, 'wave-version', input.waveVersion)
	if (input.platform) appendXmlValue(document, document.documentElement, 'platform', input.platform)
	const xml = await requestXml('/api/v1/questions', 'POST', document)
	if (!xml) throw new Error('The server returned an empty question response.')
	return parseQuestionView(xml)
}

export async function createQuestionAnswer(questionId: string, body: string): Promise<QuestionView> {
	const document = communityDocument('question-answer')
	appendXmlValue(document, document.documentElement, 'body', body)
	const xml = await requestXml(`/api/v1/questions/${encodeURIComponent(questionId)}/answers`, 'POST', document)
	if (!xml) throw new Error('The server returned an empty answer response.')
	return parseQuestionView(xml)
}

export async function voteQuestion(questionId: string, targetType: 'question' | 'answer', targetId: string, value: -1 | 0 | 1): Promise<{ score: number; viewerVote: number }> {
	const document = communityDocument('question-vote')
	appendXmlValue(document, document.documentElement, 'target-type', targetType)
	appendXmlValue(document, document.documentElement, 'target-id', targetId)
	appendXmlValue(document, document.documentElement, 'value', String(value))
	const xml = await requestXml(`/api/v1/questions/${encodeURIComponent(questionId)}/vote`, 'POST', document)
	if (!xml) throw new Error('The server returned an empty vote response.')
	return { score: Number(textOf(xml, 'score')) || 0, viewerVote: Number(textOf(xml, 'viewer-vote')) || 0 }
}

export async function acceptQuestionAnswer(questionId: string, answerId: string): Promise<QuestionView> {
	const document = communityDocument('question-accept')
	if (answerId) appendXmlValue(document, document.documentElement, 'answer-id', answerId)
	const xml = await requestXml(`/api/v1/questions/${encodeURIComponent(questionId)}/accept`, 'POST', document)
	if (!xml) throw new Error('The server returned an empty accept response.')
	return parseQuestionView(xml)
}

function parseCommunityThread(xml: XMLDocument): CommunityThread {
	const root = xml.documentElement
	const thread = Array.from(root.children).find((element) => element.localName === 'thread')
	const rootMessage = Array.from(root.children).find((element) => element.localName === 'root')
	const replies = Array.from(root.querySelector('replies')?.children ?? [])
		.filter((element) => element.localName === 'reply')
		.map((element) => parseCommunityMessage(element))
	return {
		id: thread?.getAttribute('id') ?? '',
		spaceId: thread ? childText(thread, 'space-id') : '',
		title: childText(root, 'title'),
		tags: Array.from(thread?.querySelectorAll('tags > tag') ?? []).map((tag) => tag.textContent?.trim() ?? '').filter(Boolean),
		pinned: thread ? childText(thread, 'pinned') === 'true' : false,
		locked: thread ? childText(thread, 'locked') === 'true' : false,
		root: parseCommunityMessage(rootMessage), replies,
		score: Number(childText(root, 'score')) || 0,
		viewCount: Number(childText(root, 'view-count')) || 0,
		viewerVote: Number(childText(root, 'viewer-vote')) || 0,
		subscribed: childText(root, 'subscribed') === 'true',
	}
}

export async function getSourceRepositories(): Promise<SourceRepository[]> {
  const xml = await getXml('/api/v1/source/repositories')
  return Array.from(xml.documentElement.children)
    .filter((element) => element.localName === 'repository')
    .map(parseRepository)
}

function sourceQuery(values: Record<string, string>) {
  const query = new URLSearchParams()
  for (const [key, value] of Object.entries(values)) if (value) query.set(key, value)
  const encoded = query.toString()
  return encoded ? `?${encoded}` : ''
}

export async function getSourceTree(repository: string, path = '', ref = ''): Promise<SourceTree> {
  const xml = await getXml(`/api/v1/source/repositories/${encodeURIComponent(repository)}/tree${sourceQuery({ path, ref })}`)
  const root = xml.documentElement
  const repositoryElement = Array.from(root.children).find((child) => child.localName === 'repository') as Element
  const entriesElement = Array.from(root.children).find((child) => child.localName === 'entries')
  const languagesElement = Array.from(root.children).find((child) => child.localName === 'languages')
  const readmeElement = Array.from(root.children).find((child) => child.localName === 'readme')
  const entries: SourceEntry[] = Array.from(entriesElement?.children ?? []).map((element) => ({
    name: childText(element, 'name'), path: childText(element, 'path'),
    type: childText(element, 'type') as 'tree' | 'blob', oid: childText(element, 'oid'),
    size: Number(childText(element, 'size')) || 0,
    lastCommit: parseCommit(Array.from(element.children).find((child) => child.localName === 'last-commit') ?? null),
  }))
  const languages: SourceLanguage[] = Array.from(languagesElement?.children ?? []).map((element) => {
    const name = childText(element, 'name')
    return { name, bytes: Number(childText(element, 'bytes')), files: Number(childText(element, 'files')), percent: Number(childText(element, 'percentage')), color: languageColors[name] ?? '#8b86a6' }
  })
  const readme: SourceBlob | undefined = readmeElement ? parseBlob(readmeElement as Element) : undefined
  return {
    repository: parseRepository(repositoryElement), ref: childText(root, 'ref'), path: childText(root, 'path'),
    commit: parseCommit(Array.from(root.children).find((child) => child.localName === 'commit') ?? null)!,
    entries, readme, languages,
  }
}

export async function getSourceBlob(repository: string, path: string, ref = ''): Promise<SourceBlob> {
  const xml = await getXml(`/api/v1/source/repositories/${encodeURIComponent(repository)}/blob${sourceQuery({ path, ref })}`)
  return parseBlob(xml.documentElement)
}

export async function getSourceCommits(repository: string, path = '', ref = ''): Promise<SourceCommit[]> {
  const xml = await getXml(`/api/v1/source/repositories/${encodeURIComponent(repository)}/commits${sourceQuery({ path, ref })}`)
  return Array.from(xml.documentElement.children).map((element) => parseCommit(element)!).filter(Boolean)
}

export async function getSourceCommitDetail(repository: string, oid: string): Promise<SourceCommitDetail> {
  const xml = await getXml(`/api/v1/source/repositories/${encodeURIComponent(repository)}/commits/${encodeURIComponent(oid)}`)
  const root = xml.documentElement
  const commitElement = Array.from(root.children).find((child) => child.localName === 'commit') as Element
  const filesElement = Array.from(root.children).find((child) => child.localName === 'files')
  const parentsElement = Array.from(root.children).find((child) => child.localName === 'parents')
  return {
    commit: parseCommit(commitElement)!,
    body: childContent(root, 'body').trim(),
    parents: Array.from(parentsElement?.children ?? []).map((parent) => parent.textContent?.trim() ?? '').filter(Boolean),
    files: Array.from(filesElement?.children ?? []).map((file) => ({
      status: childText(file, 'status'),
      path: childText(file, 'path'),
      oldPath: childText(file, 'old-path'),
    })),
    patch: childContent(root, 'patch'),
    patchTruncated: childText(root, 'patch-truncated') === 'true',
  }
}

export async function getSourceRefs(repository: string): Promise<SourceRefs> {
  const xml = await getXml(`/api/v1/source/repositories/${encodeURIComponent(repository)}/refs`)
  const parse = (container: string) => Array.from(xml.querySelector(container)?.children ?? []).map((element) => ({ name: childText(element, 'name'), oid: childText(element, 'oid'), updatedAt: childText(element, 'updated-at') }))
  return { branches: parse('branches'), tags: parse('tags') }
}
