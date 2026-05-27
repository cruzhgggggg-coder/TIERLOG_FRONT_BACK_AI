<?php

namespace App\Events;

use Illuminate\Broadcasting\Channel;
use Illuminate\Broadcasting\PrivateChannel;
use Illuminate\Broadcasting\InteractsWithSockets;
use Illuminate\Contracts\Broadcasting\ShouldBroadcastNow;
use Illuminate\Foundation\Events\Dispatchable;
use Illuminate\Queue\SerializesModels;

/**
 * Fired when a feedback item's status is changed (Pending→Fixed or Fixed→Validated).
 * Uses ShouldBroadcastNow to bypass the queue and send immediately via Reverb.
 */
class FeedbackStatusUpdated implements ShouldBroadcastNow
{
    use Dispatchable, InteractsWithSockets, SerializesModels;

    public int $feedbackId;
    public int $logId;
    public string $newStatus;
    public string $updatedByRole;

    /**
     * Create a new event instance.
     */
    public function __construct(
        int $feedbackId,
        int $logId,
        string $newStatus,
        string $updatedByRole,
    ) {
        $this->feedbackId    = $feedbackId;
        $this->logId         = $logId;
        $this->newStatus     = $newStatus;
        $this->updatedByRole = $updatedByRole;
    }

    /**
     * Get the channels the event should broadcast on.
     * Channel name matches what the frontend Echo subscribes to:
     * window.Echo.channel('consultation.{logId}').listen('.feedback.status-updated', ...)
     */
    public function broadcastOn(): array
    {
        return [
            new PrivateChannel('consultation.' . $this->logId),
        ];
    }

    /**
     * Custom broadcast event name (prefixed with dot on frontend).
     */
    public function broadcastAs(): string
    {
        return 'feedback.status-updated';
    }

    /**
     * Data payload sent to the frontend.
     */
    public function broadcastWith(): array
    {
        return [
            'feedback_id'     => $this->feedbackId,
            'log_id'          => $this->logId,
            'status'          => $this->newStatus,
            'updated_by_role' => $this->updatedByRole,
        ];
    }
}
