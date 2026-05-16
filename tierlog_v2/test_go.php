<?php
require 'vendor/autoload.php';
$app = require_once 'bootstrap/app.php';
$kernel = $app->make(Illuminate\Contracts\Console\Kernel::class);
$kernel->bootstrap();

use Illuminate\Support\Facades\Http;

$response = Http::get('http://localhost:8080/api/consultation?user_id=27');
echo "Status: " . $response->status() . "\n";
echo "Body: " . $response->body() . "\n";
