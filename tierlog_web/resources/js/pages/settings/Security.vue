<script setup lang="ts">
import { ref } from 'vue';
import { useAuthStore } from '@/stores/auth';
import RequireAuth from '@/components/RequireAuth.vue';
import NavBar from '@/components/NavBar.vue';
import UiPage from '@/components/UiPage.vue';
import UiHeading from '@/components/UiHeading.vue';
import UiCard from '@/components/UiCard.vue';
import UiField from '@/components/UiField.vue';
import UiButton from '@/components/UiButton.vue';
import UiBadge from '@/components/UiBadge.vue';
import { CheckCircleIcon } from '@/components/icons';

const auth = useAuthStore();

const saving = ref(false);
const currentPassword = ref('');
const newPassword = ref('');
const alertType = ref<'success' | 'error' | null>(null);
const alertMessage = ref('');

function clearAlert() {
  alertType.value = null;
  alertMessage.value = '';
}

async function updatePassword() {
  clearAlert();

  if (!currentPassword.value || !newPassword.value) {
    alertType.value = 'error';
    alertMessage.value = 'Please fill in both password fields.';
    return;
  }

  if (newPassword.value.length < 8) {
    alertType.value = 'error';
    alertMessage.value = 'New password must be at least 8 characters.';
    return;
  }

  saving.value = true;

  const res = await auth.api('/settings/password', {
    method: 'PUT',
    body: JSON.stringify({
      current_password: currentPassword.value,
      password: newPassword.value,
    }),
  });

  if (res.ok) {
    alertType.value = 'success';
    alertMessage.value = 'Password updated successfully.';
    currentPassword.value = '';
    newPassword.value = '';
  } else {
    alertType.value = 'error';
    alertMessage.value = res.error || 'Failed to update password.';
  }

  saving.value = false;
}
</script>

<template>
  <RequireAuth>
    <NavBar />
    <UiPage>
      <div class="space-y-8">
        <UiHeading title="Security" subtitle="Manage your password and credentials." title-class="font-[family-name:var(--font-display)]" />

        <div class="grid gap-6 lg:grid-cols-5">
          <div class="lg:col-span-2">
            <UiCard>
              <div class="space-y-4 p-5">
                <UiBadge text="Active" color="bg-emerald-500/10 text-emerald-400" />

                <h3 class="text-sm font-semibold text-stone-200">Password Security</h3>

                <p class="text-sm leading-relaxed text-stone-600">
                  Your password is the primary defense for your account. A strong password protects your
                  academic data and personal information.
                </p>

                <ul class="space-y-2.5">
                  <li class="flex items-start gap-2">
                    <CheckCircleIcon :size="14" color="#22c55e" class="mt-0.5 shrink-0" />
                    <span class="text-[13px] text-stone-600">Minimum 8 characters</span>
                  </li>
                  <li class="flex items-start gap-2">
                    <CheckCircleIcon :size="14" color="#22c55e" class="mt-0.5 shrink-0" />
                    <span class="text-[13px] text-stone-600">Uppercase and lowercase letters</span>
                  </li>
                  <li class="flex items-start gap-2">
                    <CheckCircleIcon :size="14" color="#22c55e" class="mt-0.5 shrink-0" />
                    <span class="text-[13px] text-stone-600">At least one number or special character</span>
                  </li>
                  <li class="flex items-start gap-2">
                    <CheckCircleIcon :size="14" color="#22c55e" class="mt-0.5 shrink-0" />
                    <span class="text-[13px] text-stone-600">Avoid reusing passwords from other services</span>
                  </li>
                </ul>
              </div>
            </UiCard>
          </div>

          <div class="lg:col-span-3">
            <UiCard>
              <div class="space-y-5 p-5 sm:p-6">
                <div class="border-b border-white/[0.04] pb-4">
                  <h3 class="text-sm font-semibold text-stone-200">Update Password</h3>
                  <p class="mt-0.5 text-xs text-stone-600">Change your account password</p>
                </div>

                <div class="space-y-4">
                  <UiField
                    v-model="currentPassword"
                    label="Current Password"
                    type="password"
                    placeholder="Enter your current password"
                  />
                  <UiField
                    v-model="newPassword"
                    label="New Password"
                    type="password"
                    placeholder="Enter a new secure password"
                  />
                </div>

                <div
                  v-if="alertType"
                  :class="[
                    'flex items-center gap-2 rounded-lg px-3.5 py-2.5 text-sm',
                    alertType === 'success'
                      ? 'border border-emerald-500/20 bg-emerald-500/5 text-emerald-400'
                      : 'border border-red-500/20 bg-red-500/5 text-red-400',
                  ]"
                >
                  {{ alertMessage }}
                </div>

                <div class="flex justify-end border-t border-white/[0.04] pt-5">
                  <UiButton
                    title="Update Password"
                    :disabled="saving"
                    @click="updatePassword"
                  />
                </div>
              </div>
            </UiCard>
          </div>
        </div>
      </div>
    </UiPage>
  </RequireAuth>
</template>
