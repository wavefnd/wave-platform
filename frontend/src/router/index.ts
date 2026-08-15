import { createRouter, createWebHistory } from 'vue-router'

import MainLayout from '../layouts/MainLayout.vue'
import CommunityPage from '../pages/CommunityPage.vue'
import DocsPage from '../pages/DocsPage.vue'
import HomePage from '../pages/HomePage.vue'
import BlogPage from '../pages/BlogPage.vue'
import MailPage from '../pages/MailPage.vue'
import NotFoundPage from '../pages/NotFoundPage.vue'
import QuestionsPage from '../pages/QuestionsPage.vue'
import ReleasePage from '../pages/ReleasePage.vue'
import SearchPage from '../pages/SearchPage.vue'
import SourcePage from '../pages/SourcePage.vue'
import { useAuthStore } from '../stores/auth'

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
		{ path: 'blog/:slug', name: 'blog-post', component: BlogPage },
        { path: 'docs', name: 'docs', component: DocsPage },
        { path: 'docs/:pathMatch(.*)*', name: 'document', component: DocsPage },
        { path: 'mail', name: 'mail', component: MailPage },
        { path: 'community', name: 'community', component: CommunityPage },
		{ path: 'community/new', name: 'community-new', component: CommunityPage },
		{ path: 'community/showcase', name: 'community-showcase', component: CommunityPage, meta: { showcase: true } },
		{ path: 'community/showcase/new', name: 'community-showcase-new', component: CommunityPage, meta: { showcase: true } },
		{ path: 'community/showcase/:thread', name: 'community-showcase-thread', component: CommunityPage, meta: { showcase: true } },
        { path: 'community/thread/:thread', name: 'community-thread', component: CommunityPage },
        { path: 'lunastev', name: 'personal-space', component: CommunityPage, meta: { personalSpace: true } },
        { path: 'lunastev/new', name: 'personal-space-new', component: CommunityPage, meta: { personalSpace: true } },
        { path: 'lunastev/thread/:thread', name: 'personal-space-thread', component: CommunityPage, meta: { personalSpace: true } },
        { path: 'releases/:slug', name: 'release', component: ReleasePage },
        { path: 'community/announcements/:slug', redirect: (to) => `/releases/${String(to.params.slug)}` },
        { path: 'questions', name: 'questions', component: QuestionsPage },
        { path: 'questions/new', name: 'question-new', component: QuestionsPage },
        { path: 'questions/:question', name: 'question-detail', component: QuestionsPage },
        { path: 'source', name: 'source', component: SourcePage },
        { path: 'source/:repository', name: 'source-repository', component: SourcePage },
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
		{ path: 'admin/:section(blog|accounts|mailbox|mail-queue|git-mirrors|audit-log|security|modules|system)', name: 'admin-section', component: () => import('../pages/AdminPage.vue') },
        { path: ':pathMatch(.*)*', name: 'not-found', component: NotFoundPage },
      ],
    },
  ],
})

router.beforeEach(async (to) => {
  if (!String(to.name ?? '').startsWith('admin') && !to.meta.requiresAuth) return true
  const auth = useAuthStore()
  await auth.initialize()
	if (to.meta.requiresAuth && auth.account) return true
	if (String(to.name ?? '').startsWith('admin') && (auth.account?.owner || auth.account?.administrator)) return true
  return { name: 'login', query: { redirect: to.fullPath } }
})

export default router
