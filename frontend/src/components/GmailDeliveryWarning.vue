<script setup lang="ts">
import { TriangleAlert } from '@lucide/vue'

import { useI18n } from '../i18n'

const { t } = useI18n()
</script>

<template>
	<p class="gmail-delivery-warning" role="alert">
		<TriangleAlert :size="17" aria-hidden="true" />
		<span>{{ t('mail.gmailWarning') }} <a href="https://proton.me/mail" target="_blank" rel="noopener noreferrer">Proton Mail</a>.</span>
	</p>
</template>
