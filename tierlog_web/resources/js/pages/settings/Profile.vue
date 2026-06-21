<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useAuthStore } from '@/stores/auth';
import RequireAuth from '@/components/RequireAuth.vue';
import NavBar from '@/components/NavBar.vue';
import UiPage from '@/components/UiPage.vue';
import UiHeading from '@/components/UiHeading.vue';
import UiCard from '@/components/UiCard.vue';
import UiField from '@/components/UiField.vue';
import UiButton from '@/components/UiButton.vue';

const auth = useAuthStore();

const saving = ref(false);
const alertType = ref<'success' | 'error' | null>(null);
const alertMessage = ref('');

const name = ref('');
const email = ref('');
const nim = ref('');
const prodi = ref('');
const thesisTitle = ref('');
const lecturerId = ref('');
const nip = ref('');
const faculty = ref('');
const expertise = ref('');
const aiConstraints = ref('');

const isStudent = computed(() => auth.user?.role === 'student');
const isLecturer = computed(() => auth.user?.role === 'lecturer');

function populateFromUser() {
  const u = auth.user;
  if (!u) return;
  name.value = u.name;
  email.value = u.email;
  if (u.student) {
    nim.value = u.student.nim ?? '';
    prodi.value = u.student.prodi ?? '';
    thesisTitle.value = u.student.thesis_title ?? '';
    lecturerId.value = u.student.lecturer_id ? String(u.student.lecturer_id) : '';
  }
  if (u.lecturer) {
    nip.value = u.lecturer.nip ?? '';
    faculty.value = u.lecturer.faculty ?? '';
    expertise.value = u.lecturer.keahlian ?? '';
    aiConstraints.value = u.lecturer.ai_constraints ?? '';
  }
}

onMounted(() => {
  populateFromUser();
});

function clearAlert() {
  alertType.value = null;
  alertMessage.value = '';
}

async function saveProfile() {
  clearAlert();
  saving.value = true;

  const body: Record<string, unknown> = {
    name: name.value,
    email: email.value,
  };

  if (isStudent.value) {
    body.nim = nim.value;
    body.prodi = prodi.value;
    body.thesis_title = thesisTitle.value;
    body.lecturer_id = lecturerId.value || null;
  }

  if (isLecturer.value) {
    body.nip = nip.value;
    body.faculty = faculty.value;
    body.keahlian = expertise.value;
    body.ai_constraints = aiConstraints.value;
  }

  const res = await auth.api<{ user: typeof auth.user }>('/settings/profile', {
    method: 'PATCH',
    body: JSON.stringify(body),
  });

  if (res.ok && res.data?.user) {
    auth.user = res.data.user;
    alertType.value = 'success';
    alertMessage.value = 'Profile updated successfully.';
    populateFromUser();
  } else {
    alertType.value = 'error';
    alertMessage.value = res.error || 'Failed to update profile.';
  }

  saving.value = false;
}
</script>

<template>
  <RequireAuth>
    <NavBar />
    <UiPage>
      <div class="space-y-8">
        <UiHeading title="Profile Settings" subtitle="Manage your account and academic information." title-class="font-[family-name:var(--font-display)]" />

        <UiCard>
          <div class="space-y-6 p-5 sm:p-6">
            <div class="border-b border-white/[0.04] pb-4">
              <h3 class="text-sm font-semibold text-stone-200">Account</h3>
              <p class="mt-0.5 text-xs text-stone-600">Basic account details</p>
            </div>

            <div class="grid gap-4 sm:grid-cols-2">
              <UiField v-model="name" label="Full Name" placeholder="Enter your name" />
              <UiField v-model="email" label="Email Address" type="email" placeholder="you@example.com" />
            </div>

            <template v-if="isStudent">
              <div class="border-t border-white/[0.04] pt-5">
                <div class="mb-4">
                  <h3 class="text-sm font-semibold text-stone-200">Student Information</h3>
                  <p class="mt-0.5 text-xs text-stone-600">Academic records and affiliation</p>
                </div>
                <div class="grid gap-4 sm:grid-cols-2">
                  <UiField v-model="nim" label="NIM" placeholder="Student ID number" />
                  <UiField v-model="prodi" label="Prodi" placeholder="Study program" />
                  <UiField v-model="thesisTitle" label="Thesis Title" placeholder="Your thesis title" class="sm:col-span-2" />
                  <UiField v-model="lecturerId" label="Lecturer ID" placeholder="Assigned lecturer ID" />
                </div>
              </div>
            </template>

            <template v-if="isLecturer">
              <div class="border-t border-white/[0.04] pt-5">
                <div class="mb-4">
                  <h3 class="text-sm font-semibold text-stone-200">Lecturer Information</h3>
                  <p class="mt-0.5 text-xs text-stone-600">Academic credentials and expertise</p>
                </div>
                <div class="grid gap-4 sm:grid-cols-2">
                  <UiField v-model="nip" label="NIP" placeholder="Staff ID number" />
                  <UiField v-model="faculty" label="Faculty" placeholder="Faculty name" />
                  <UiField v-model="expertise" label="Expertise" placeholder="Area of expertise" />
                  <UiField v-model="aiConstraints" label="AI Guidelines & Constraints" type="textarea" placeholder="Modify system prompt / instructions for AI classification (e.g. 'Must focus on structural logic rather than grammar')" class="sm:col-span-2" />
                </div>
              </div>
            </template>

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
                title="Save Changes"
                :disabled="saving"
                @click="saveProfile"
              />
            </div>
          </div>
        </UiCard>
      </div>
    </UiPage>
  </RequireAuth>
</template>
