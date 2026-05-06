<script setup lang="ts">
import { Form, Head, Link, usePage } from '@inertiajs/vue3';
import { computed } from 'vue';
import ProfileController from '@/actions/App/Http/Controllers/Settings/ProfileController';
import DeleteUser from '@/components/DeleteUser.vue';
import Heading from '@/components/Heading.vue';
import InputError from '@/components/InputError.vue';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { edit } from '@/routes/profile';
import { send } from '@/routes/verification';

type Props = {
    mustVerifyEmail: boolean;
    status?: string;
    lecturers: any[];
};

defineProps<Props>();

defineOptions({
    layout: {
        breadcrumbs: [
            {
                title: 'Profile settings',
                href: edit(),
            },
        ],
    },
});

const page = usePage();
const user = computed(() => page.props.auth.user as any);
</script>

<template>
    <Head title="Profile settings" />

    <h1 class="sr-only">Profile settings</h1>

    <div class="flex flex-col space-y-6">
        <Heading
            variant="small"
            title="Profile information"
            description="Update your name and email address"
        />

        <Form
            v-bind="ProfileController.update.form()"
            class="space-y-6"
            v-slot="{ errors, processing }"
        >
            <div class="grid gap-2">
                <Label for="name">Name</Label>
                <Input
                    id="name"
                    class="mt-1 block w-full"
                    name="name"
                    :default-value="user.name"
                    required
                    autocomplete="name"
                    placeholder="Full name"
                />
                <InputError class="mt-2" :message="errors.name" />
            </div>

            <div class="grid gap-2">
                <Label for="email">Email address</Label>
                <Input
                    id="email"
                    type="email"
                    class="mt-1 block w-full"
                    name="email"
                    :default-value="user.email"
                    required
                    autocomplete="username"
                    placeholder="Email address"
                />
                <InputError class="mt-2" :message="errors.email" />
            </div>

            <!-- Student Specific Fields -->
            <template v-if="user.role === 'student'">
                <div class="grid gap-2">
                    <Label for="nim">NIM</Label>
                    <Input
                        id="nim"
                        class="mt-1 block w-full"
                        name="nim"
                        :default-value="user.student?.nim"
                        required
                        placeholder="Nomor Induk Mahasiswa"
                    />
                    <InputError class="mt-2" :message="errors.nim" />
                </div>

                <div class="grid gap-2">
                    <Label for="prodi">Program Studi</Label>
                    <Input
                        id="prodi"
                        class="mt-1 block w-full"
                        name="prodi"
                        :default-value="user.student?.prodi"
                        required
                        placeholder="Contoh: Teknik Informatika"
                    />
                    <InputError class="mt-2" :message="errors.prodi" />
                </div>

                <div class="grid gap-2">
                    <Label for="thesis_title">Judul Tugas Akhir</Label>
                    <Input
                        id="thesis_title"
                        class="mt-1 block w-full"
                        name="thesis_title"
                        :default-value="user.student?.thesis_title"
                        placeholder="Masukkan judul TA lengkap"
                    />
                    <InputError class="mt-2" :message="errors.thesis_title" />
                </div>

                <div class="grid gap-2">
                    <Label for="lecturer_id">Dosen Pembimbing</Label>
                    <select
                        id="lecturer_id"
                        name="lecturer_id"
                        class="mt-1 block w-full bg-background border border-input rounded-md px-3 py-2 text-sm"
                        :value="user.student?.lecturer_id"
                    >
                        <option value="">Pilih Dosen Pembimbing</option>
                        <option v-for="lecturer in lecturers" :key="lecturer.id" :value="lecturer.id">
                            {{ lecturer.name }}
                        </option>
                    </select>
                    <InputError class="mt-2" :message="errors.lecturer_id" />
                </div>
            </template>

            <!-- Lecturer Specific Fields -->
            <template v-if="user.role === 'lecturer'">
                <div class="grid gap-2">
                    <Label for="nip">NIP</Label>
                    <Input
                        id="nip"
                        class="mt-1 block w-full"
                        name="nip"
                        :default-value="user.lecturer?.nip"
                        required
                        placeholder="Nomor Induk Pegawai"
                    />
                    <InputError class="mt-2" :message="errors.nip" />
                </div>

                <div class="grid gap-2">
                    <Label for="faculty">Fakultas</Label>
                    <Input
                        id="faculty"
                        class="mt-1 block w-full"
                        name="faculty"
                        :default-value="user.lecturer?.faculty"
                        required
                        placeholder="Contoh: Fakultas Teknik"
                    />
                    <InputError class="mt-2" :message="errors.faculty" />
                </div>
            </template>

            <div v-if="mustVerifyEmail && !user.email_verified_at">
                <p class="-mt-4 text-sm text-muted-foreground">
                    Your email address is unverified.
                    <Link
                        :href="send()"
                        as="button"
                        class="text-foreground underline decoration-neutral-300 underline-offset-4 transition-colors duration-300 ease-out hover:decoration-current! dark:decoration-neutral-500"
                    >
                        Click here to resend the verification email.
                    </Link>
                </p>

                <div
                    v-if="status === 'verification-link-sent'"
                    class="mt-2 text-sm font-medium text-green-600"
                >
                    A new verification link has been sent to your email address.
                </div>
            </div>

            <div class="flex items-center gap-4">
                <Button :disabled="processing" data-test="update-profile-button"
                    >Save</Button
                >
            </div>
        </Form>
    </div>

    <DeleteUser />
</template>
