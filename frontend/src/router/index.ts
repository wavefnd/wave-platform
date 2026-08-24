import { createRouter, createWebHistory } from 'vue-router'

import MainLayout from '../layouts/MainLayout.vue'
import CommunityPage from '../pages/CommunityPage.vue'
import DocsPage from '../pages/DocsPage.vue'
import HomePage from '../pages/HomePage.vue'
import BlogPage from '../pages/BlogPage.vue'
import BlogEditorPage from '../pages/BlogEditorPage.vue'
import MailPage from '../pages/MailPage.vue'
import NotFoundPage from '../pages/NotFoundPage.vue'
import QuestionsPage from '../pages/QuestionsPage.vue'
import RFCPage from '../pages/RFCPage.vue'
import PatchesPage from '../pages/PatchesPage.vue'
import SearchPage from '../pages/SearchPage.vue'
import SourcePage from '../pages/SourcePage.vue'
import { useAuthStore } from '../stores/auth'
import { initialDocumentLocale } from '../services/documentLocale'

const router = createRouter({
  history: createWebHistory(),
  scrollBehavior: () => ({ top: 0 }),
  routes: [
    {
      path: '/',
      component: MainLayout,
      children: [
        { path: '', name: 'home', component: HomePage },
		{ path: 'blog', name: 'blog', component: BlogPage },
		{ path: 'blog/editor', name: 'blog-editor', component: BlogEditorPage, meta: { requiresAuth: true, requiresAdmin: true } },
		{ path: 'blog/editor/:slug', name: 'blog-editor-post', component: BlogEditorPage, meta: { requiresAuth: true, requiresAdmin: true } },
		{ path: 'blog/:slug', name: 'blog-post', component: BlogPage },
		{ path: 'releases', name: 'releases', component: BlogPage, meta: { blogCategory: 'release' } },
		{ path: 'releases/:slug', name: 'release-detail', component: BlogPage, meta: { blogCategory: 'release' } },
        { path: 'docs', name: 'docs', redirect: () => ({ name: 'docs-locale', params: { docLocale: initialDocumentLocale() } }) },
        { path: 'docs/:docLocale(en|ko|ja|zh|es|de|ru|id|vi)', name: 'docs-locale', component: DocsPage },
        { path: 'docs/:docLocale(en|ko|ja|zh|es|de|ru|id|vi)/:pathMatch(.*)*', name: 'document', component: DocsPage },
        { path: 'docs/:pathMatch(.*)*', name: 'document-legacy', redirect: (to) => ({
          name: 'document', params: { docLocale: initialDocumentLocale(), pathMatch: to.params.pathMatch },
        }) },
        { path: 'mail', name: 'mail', component: MailPage },
		{ path: 'mail/personal', name: 'mail-personal', component: MailPage },
		{ path: 'mail/lists', name: 'mail-lists', component: MailPage },
		{ path: 'mail/lists/:list', name: 'mail-list', component: MailPage },
		{ path: 'mail/lists/:list/thread/:thread', name: 'mail-list-thread', component: MailPage },
		{ path: 'mail/team', name: 'mail-team', component: MailPage, meta: { requiresAuth: true, requiresAdmin: true } },
		{ path: 'mail/lists/patchs/reviews', name: 'patch-reviews', component: PatchesPage, meta: { requiresAuth: true } },
		{ path: 'mail/lists/patchs/patch/:patch', name: 'patch-detail', component: PatchesPage, meta: { requiresAuth: true } },
        { path: 'community', name: 'community', component: CommunityPage },
		{ path: 'community/new', name: 'community-new', component: CommunityPage },
		{ path: 'community/showcase', name: 'community-showcase', component: CommunityPage, meta: { showcase: true } },
		{ path: 'community/showcase/new', name: 'community-showcase-new', component: CommunityPage, meta: { showcase: true } },
		{ path: 'community/showcase/:thread', name: 'community-showcase-thread', component: CommunityPage, meta: { showcase: true } },
        { path: 'community/thread/:thread', name: 'community-thread', component: CommunityPage },
        { path: 'lunastev', name: 'personal-space', component: CommunityPage, meta: { personalSpace: true } },
        { path: 'lunastev/new', name: 'personal-space-new', component: CommunityPage, meta: { personalSpace: true } },
        { path: 'lunastev/thread/:thread', name: 'personal-space-thread', component: CommunityPage, meta: { personalSpace: true } },
		{ path: 'community/announcements/:slug', redirect: (to) => `/blog/${String(to.params.slug)}` },
        { path: 'questions', name: 'questions', component: QuestionsPage },
        { path: 'questions/new', name: 'question-new', component: QuestionsPage },
        { path: 'questions/:question', name: 'question-detail', component: QuestionsPage },
		{ path: 'rfcs', name: 'rfcs', component: RFCPage },
		{ path: 'rfcs/new', name: 'rfc-new', component: RFCPage, meta: { requiresAuth: true } },
		{ path: 'rfcs/:number', name: 'rfc-detail', component: RFCPage },
		{ path: 'rfcs/:number/edit', name: 'rfc-edit', component: RFCPage, meta: { requiresAuth: true } },
        { path: 'source', name: 'source', component: SourcePage },
        { path: 'source/:repository', name: 'source-repository', component: SourcePage },
        { path: 'patches', redirect: (to) => ({ name: 'patch-reviews', query: to.query, hash: to.hash }) },
        { path: 'patches/:patch', redirect: (to) => ({ name: 'patch-detail', params: { patch: to.params.patch }, query: to.query, hash: to.hash }) },
        { path: 'search', name: 'search', component: SearchPage },
		{ path: 'user', name: 'user-directory', component: () => import('../pages/UserPage.vue') },
		{ path: 'user/id/:account', name: 'user-id-profile', component: () => import('../pages/UserPage.vue') },
		{ path: 'user/:user', name: 'user-profile', component: () => import('../pages/UserPage.vue') },
		{ path: 'login', name: 'login', component: () => import('../pages/LoginPage.vue') },
		{ path: 'register', name: 'register', component: () => import('../pages/RegisterPage.vue') },
		{ path: 'account/recover', name: 'account-recover', component: () => import('../pages/RecoveryPage.vue') },
		{ path: 'account/verify-recovery', name: 'verify-recovery', component: () => import('../pages/VerifyRecoveryEmailPage.vue') },
		{ path: 'account/security', name: 'account-security', component: () => import('../pages/SecuritySettingsPage.vue'), meta: { requiresAuth: true } },
        { path: 'admin', name: 'admin', component: () => import('../pages/AdminPage.vue'), meta: { adminSection: 'overview' } },
		{ path: 'admin/:section(webhooks|accounts|mailbox|mail-queue|git-mirrors|audit-log|security|modules|system)', name: 'admin-section', component: () => import('../pages/AdminPage.vue') },
        { path: ':pathMatch(.*)*', name: 'not-found', component: NotFoundPage },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  const isAdminRoute = String(to.name ?? '').startsWith('admin')
  if (!isAdminRoute && !to.meta.requiresAuth && !to.meta.requiresAdmin) return true
  const auth = useAuthStore()
  await auth.initialize()
  const isAdmin = Boolean(auth.account?.owner || auth.account?.administrator)
	if (to.meta.requiresAdmin && isAdmin) return true
	if (isAdminRoute && isAdmin) return true
	if (to.meta.requiresAuth && !to.meta.requiresAdmin && auth.account) return true
  return { name: 'login', query: { redirect: to.fullPath } }
})

export default router
