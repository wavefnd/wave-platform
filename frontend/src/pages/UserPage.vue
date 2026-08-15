<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useI18n } from '../i18n'
import { getUser, getUserByID, getUsers, updateUserProfile, updateWaveAddress, type UserActivity, type UserProfile } from '../services/http'
import { useAuthStore } from '../stores/auth'
import UiInlineState from '../ui/UiInlineState.vue'

const { locale, t } = useI18n()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const users = ref<UserProfile[]>([])
const profile = ref<UserProfile | null>(null)
const loading = ref(false)
const error = ref('')
const notice = ref('')
const editing = ref(false)
const displayName = ref('')
const bio = ref('')
const localPart = ref('')
const code = ref('')
const timeZone = ref('UTC')
const timeZones = ['UTC', 'Asia/Seoul', 'Asia/Tokyo', 'Asia/Singapore', 'Europe/London', 'Europe/Paris', 'America/New_York', 'America/Los_Angeles']
const filter = ref<'all' | UserActivity['kind']>('all')
const activityKinds = ['all', 'community-post', 'community-comment', 'question', 'answer'] as const

const directoryMode = computed(() => route.name === 'user-directory')
const ownProfile = computed(() => Boolean(profile.value && auth.account?.email.toLowerCase() === profile.value.email.toLowerCase()))
const visibleActivities = computed(() => profile.value?.activities.filter((item) => filter.value === 'all' || item.kind === filter.value) ?? [])
function setFilter(kind: typeof activityKinds[number]) { filter.value = kind }

function formatDate(value: string) {
  if (!value) return ''
  return new Intl.DateTimeFormat(locale.value === 'ko' ? 'ko-KR' : 'en-US', { dateStyle: 'medium', timeStyle: 'short', timeZone: auth.account?.timeZone || timeZone.value }).format(new Date(value))
}

async function load() {
  loading.value = true; error.value = ''; notice.value = ''; editing.value = false
  try {
    await auth.initialize()
    if (directoryMode.value) { users.value = await getUsers(); profile.value = null }
    else {
      profile.value = route.name === 'user-id-profile'
        ? await getUserByID(String(route.params.account ?? ''))
        : await getUser(String(route.params.user ?? ''))
      displayName.value = profile.value.displayName; bio.value = profile.value.bio; timeZone.value = profile.value.timeZone
      localPart.value = profile.value.email.split('@')[0] ?? ''
    }
  } catch (reason) { error.value = reason instanceof Error ? reason.message : t('common.loadError') }
  finally { loading.value = false }
}

async function saveProfile() {
  error.value = ''; notice.value = ''
  try {
    profile.value = await updateUserProfile(displayName.value, bio.value, timeZone.value)
    await auth.initialize(true); notice.value = t('user.profileSaved'); editing.value = false
  } catch (reason) { error.value = reason instanceof Error ? reason.message : t('user.updateFailed') }
}

async function saveAddress() {
  error.value = ''; notice.value = ''
  try {
    const updated = await updateWaveAddress(localPart.value, code.value)
    profile.value = updated; code.value = ''; await auth.initialize(true)
    notice.value = t('user.addressSaved')
    await router.replace({ name: 'user-profile', params: { user: updated.username } })
  } catch (reason) { error.value = reason instanceof Error ? reason.message : t('user.updateFailed') }
}

onMounted(load)
watch(() => route.fullPath, load)
</script>

<template>
  <main class="user-page portal-width">
    <header class="user-page-heading">
      <div><span>{{ t('user.directory') }}</span><h1>{{ directoryMode ? t('user.people') : profile?.displayName }}</h1></div>
      <RouterLink v-if="!directoryMode" to="/user">{{ t('user.allUsers') }}</RouterLink>
    </header>
    <UiInlineState v-if="loading" :message="t('common.loading')" />
    <UiInlineState v-else-if="error && !profile" :message="error" />

    <section v-else-if="directoryMode" class="user-directory" aria-live="polite">
      <RouterLink v-for="item in users" :key="item.email" :to="`/user/${encodeURIComponent(item.username)}`">
        <span class="user-avatar">{{ Array.from(item.displayName)[0]?.toUpperCase() ?? 'W' }}</span>
        <span><strong>{{ item.displayName }}</strong><small>{{ item.email }}</small><p v-if="item.bio">{{ item.bio }}</p></span>
      </RouterLink>
    </section>

    <template v-else-if="profile">
      <section class="user-profile-header">
        <span class="user-avatar large">{{ Array.from(profile.displayName)[0]?.toUpperCase() ?? 'W' }}</span>
        <div><h2>{{ profile.displayName }}</h2><div class="user-profile-meta"><a :href="`mailto:${profile.email}`">{{ profile.email }}</a><small>{{ t('user.joined') }} {{ formatDate(profile.joinedAt) }}</small></div><p v-if="profile.bio">{{ profile.bio }}</p></div>
        <button v-if="ownProfile" class="ui-button" type="button" @click="editing = !editing">{{ t('user.editProfile') }}</button>
      </section>

      <div v-if="notice" class="form-notice" role="status">{{ notice }}</div>
      <div v-if="error" class="form-notice error" role="alert">{{ error }}</div>

      <section v-if="ownProfile && editing" class="user-settings">
        <form @submit.prevent="saveProfile">
          <h2>{{ t('user.profileSettings') }}</h2>
          <label>{{ t('auth.displayName') }}<input v-model="displayName" required maxlength="80" /></label>
          <label>{{ t('user.bio') }}<textarea v-model="bio" maxlength="500" rows="5" /></label>
		  <label>{{ t('user.timeZone') }}<select v-model="timeZone"><option v-for="zone in timeZones" :key="zone" :value="zone">{{ zone }}</option></select></label>
          <button class="ui-button primary" type="submit">{{ t('common.save') }}</button>
        </form>
        <form v-if="profile.addressChoiceAllowed" @submit.prevent="saveAddress">
          <h2>{{ t('user.waveAddress') }}</h2>
          <p>{{ t('user.addressIdentityHelp') }}</p>
          <label>{{ t('user.localPart') }}<input v-model="localPart" required maxlength="60" /></label>
          <label>{{ t('auth.currentCode') }}<input v-model="code" required inputmode="numeric" pattern="[0-9]{6}" maxlength="6" autocomplete="one-time-code" /></label>
          <button class="ui-button primary" type="submit">{{ t('user.changeAddress') }}</button>
        </form>
      </section>

      <section class="user-activity">
        <header><h2>{{ t('user.activity') }}</h2><nav class="ui-tabs" :aria-label="t('user.activity')">
          <button v-for="kind in activityKinds" :key="kind" type="button" :class="{ active: filter === kind }" @click="setFilter(kind)">{{ t(`user.activity.${kind}`) }}</button>
        </nav></header>
        <RouterLink v-for="item in visibleActivities" :key="`${item.kind}-${item.url}-${item.createdAt}`" :to="item.url" class="user-activity-row">
          <span>{{ t(`user.activity.${item.kind}`) }}</span><strong>{{ item.title }}</strong><p>{{ item.excerpt }}</p><time :datetime="item.createdAt">{{ formatDate(item.createdAt) }}</time>
        </RouterLink>
        <p v-if="visibleActivities.length === 0" class="compact-empty">{{ t('user.noActivity') }}</p>
      </section>
    </template>
  </main>
</template>
