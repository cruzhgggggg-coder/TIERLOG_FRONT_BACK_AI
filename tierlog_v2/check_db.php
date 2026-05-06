<?php
require 'vendor/autoload.php';
$app = require_once 'bootstrap/app.php';
$app->make(Illuminate\Contracts\Console\Kernel::class)->bootstrap();
$lecturers = \App\Models\Lecturer::all();
echo json_encode($lecturers);
