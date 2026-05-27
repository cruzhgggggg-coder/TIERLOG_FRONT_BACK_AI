<?php

namespace App\Events;

use Illuminate\Broadcasting\Channel;
use Illuminate\Broadcasting\PrivateChannel;
use Illuminate\Broadcasting\InteractsWithSockets;
use Illuminate\Contracts\Broadcasting\ShouldBroadcastNow;
use Illuminate\Foundation\Events\Dispatchable;
use Illuminate\Queue\SerializesModels;

class AIChatMessageBroadcasted implements ShouldBroadcastNow
{
    use Dispatchable, InteractsWithSockets, SerializesModels;

    public int $logId;
    public string $role; // 'user' (student) or 'ai' (oracle)
    public string $content;

    /**
     * Create a new event instance.
     */
    public function __construct(int $logId, string $role, string $content)
    {
        $this->logId = $logId;
        $this->role = $role;
        $this->content = $content;
    }

    /**
     * Get the channels the event should broadcast on.
     */
    public function broadcastOn(): array
    {
        return [
            new PrivateChannel('consultation.' . $this->logId),
        ];
    }

    /**
     * Get the broadcast name.
     */
    public function broadcastAs(): string
    {
        return 'chat.message';
    }
}
