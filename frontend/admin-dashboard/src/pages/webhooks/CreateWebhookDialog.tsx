/**
 * Create Webhook Dialog
 * 
 * AUTHORITY MODEL:
 * Who: SYSTEM users (system webhooks) OR TENANT users (tenant webhooks)
 * Scope: System-wide OR Tenant-scoped
 * Permission: webhooks:create
 * 
 * SECURITY:
 * - One-time signing secret display
 * - Copy confirmation before close
 * - Clear warnings about sensitive data
 * 
 * UI CONTRACT MODE:
 * - createWebhook throws APINotConnectedError
 */

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Copy, CheckCircle2, AlertTriangle, Webhook } from 'lucide-react';
import { APINotConnectedError, getAPINotConnectedMessage, isAPINotConnected } from '@/lib/errors';

const createWebhookSchema = z.object({
    name: z.string().min(1, 'Webhook name is required'),
    url: z.string().url('Must be a valid HTTPS URL').startsWith('https://', 'URL must use HTTPS'),
    events: z.array(z.string()).min(1, 'At least one event is required'),
    retry_policy: z.object({
        max_attempts: z.number().min(1).max(10),
        backoff_seconds: z.number().min(1).max(3600),
    }).optional(),
});

type CreateWebhookFormData = z.infer<typeof createWebhookSchema>;

interface CreateWebhookDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    scope: 'system' | 'tenant';
    tenantId?: string;
}

const AVAILABLE_EVENTS = [
    { value: 'user.created', label: 'User Created', description: 'When a new user is created' },
    { value: 'user.updated', label: 'User Updated', description: 'When a user is modified' },
    { value: 'user.deleted', label: 'User Deleted', description: 'When a user is deleted' },
    { value: 'user.suspended', label: 'User Suspended', description: 'When a user is suspended' },
    { value: 'role.assigned', label: 'Role Assigned', description: 'When a role is assigned to a user' },
    { value: 'role.revoked', label: 'Role Revoked', description: 'When a role is revoked from a user' },
    { value: 'session.created', label: 'Session Created', description: 'When a user logs in' },
    { value: 'session.revoked', label: 'Session Revoked', description: 'When a session is revoked' },
];

export function CreateWebhookDialog({ open, onOpenChange, scope, tenantId }: CreateWebhookDialogProps) {
    const queryClient = useQueryClient();
    const [error, setError] = useState<string | null>(null);
    const [createdWebhook, setCreatedWebhook] = useState<{ signing_secret: string; name: string } | null>(null);
    const [secretCopied, setSecretCopied] = useState(false);

    const {
        register,
        handleSubmit,
        formState: { errors },
        reset,
        watch,
        setValue,
    } = useForm<CreateWebhookFormData>({
        resolver: zodResolver(createWebhookSchema),
        defaultValues: {
            events: [],
            retry_policy: {
                max_attempts: 3,
                backoff_seconds: 60,
            },
        },
    });

    const selectedEvents = watch('events') || [];

    const createWebhookMutation = useMutation({
        mutationFn: async (data: CreateWebhookFormData) => {
            // UI CONTRACT MODE: Throw APINotConnectedError
            throw new APINotConnectedError('webhooks.create');
        },
        onSuccess: (data: any) => {
            queryClient.invalidateQueries({ queryKey: scope === 'system' ? ['system-webhooks'] : ['webhooks', tenantId] });
            setCreatedWebhook(data);
            setError(null);
        },
        onError: (err: any) => {
            setError(getAPINotConnectedMessage(err));
        },
    });

    const onSubmit = (data: CreateWebhookFormData) => {
        setError(null);
        createWebhookMutation.mutate(data);
    };

    const handleClose = () => {
        if (createdWebhook && !secretCopied) {
            if (!confirm('You have not copied the signing secret. It will not be shown again. Are you sure you want to close?')) {
                return;
            }
        }
        onOpenChange(false);
        reset();
        setCreatedWebhook(null);
        setSecretCopied(false);
        setError(null);
    };

    const copySecret = () => {
        if (createdWebhook) {
            navigator.clipboard.writeText(createdWebhook.signing_secret);
            setSecretCopied(true);
        }
    };

    // Show success screen after creation
    if (createdWebhook) {
        return (
            <Dialog open={open} onOpenChange={handleClose}>
                <DialogContent className="max-w-md">
                    <DialogHeader>
                        <DialogTitle>Webhook Created</DialogTitle>
                        <DialogDescription>
                            Save the signing secret securely. It will not be shown again.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="space-y-4">
                        <Alert className="bg-green-50 border-green-200">
                            <CheckCircle2 className="h-4 w-4 text-green-600" />
                            <AlertDescription className="text-green-800">
                                Webhook "{createdWebhook.name}" created successfully
                            </AlertDescription>
                        </Alert>

                        <div className="space-y-2">
                            <Label>Signing Secret</Label>
                            <div className="flex gap-2">
                                <Input
                                    value={createdWebhook.signing_secret}
                                    readOnly
                                    className="font-mono text-sm"
                                    type="password"
                                />
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={copySecret}
                                    className={secretCopied ? 'bg-green-50' : ''}
                                >
                                    {secretCopied ? (
                                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                                    ) : (
                                        <Copy className="h-4 w-4" />
                                    )}
                                </Button>
                            </div>
                            <p className="text-xs text-red-600">
                                ⚠️ This secret will not be shown again. Use it to verify webhook signatures.
                            </p>
                        </div>

                        <Alert className="bg-orange-50 border-orange-200">
                            <Webhook className="h-4 w-4 text-orange-600" />
                            <AlertDescription className="text-orange-800 text-sm">
                                <strong>Security:</strong> Use this secret to verify webhook signatures and prevent spoofing.
                            </AlertDescription>
                        </Alert>
                    </div>

                    <DialogFooter>
                        <Button onClick={handleClose} variant={secretCopied ? 'default' : 'outline'}>
                            {secretCopied ? 'Done' : 'I have saved the secret'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        );
    }

    return (
        <Dialog open={open} onOpenChange={handleClose}>
            <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Create Webhook</DialogTitle>
                    <DialogDescription>
                        {scope === 'system'
                            ? 'Create a system-wide webhook for all tenants'
                            : 'Create a webhook for this tenant'}
                    </DialogDescription>
                </DialogHeader>

                <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
                    {error && (
                        <Alert variant={isAPINotConnected(error) ? 'default' : 'destructive'} className={isAPINotConnected(error) ? 'bg-blue-50 border-blue-200' : ''}>
                            <AlertDescription className={isAPINotConnected(error) ? 'text-blue-800' : ''}>
                                {error}
                            </AlertDescription>
                        </Alert>
                    )}

                    {/* Scope Badge */}
                    <Badge className={scope === 'system' ? 'bg-blue-100 text-blue-800' : 'bg-green-100 text-green-800'}>
                        {scope === 'system' ? 'System Webhook (All Tenants)' : 'Tenant Webhook'}
                    </Badge>

                    {/* Webhook Name */}
                    <div className="space-y-2">
                        <Label htmlFor="name">Webhook Name *</Label>
                        <Input
                            id="name"
                            {...register('name')}
                            placeholder="Production Notifications"
                            disabled={createWebhookMutation.isPending}
                        />
                        {errors.name && (
                            <p className="text-sm text-red-600">{errors.name.message}</p>
                        )}
                    </div>

                    {/* URL */}
                    <div className="space-y-2">
                        <Label htmlFor="url">Webhook URL *</Label>
                        <Input
                            id="url"
                            {...register('url')}
                            placeholder="https://example.com/webhooks/arauth"
                            disabled={createWebhookMutation.isPending}
                            className="font-mono text-sm"
                        />
                        {errors.url && (
                            <p className="text-sm text-red-600">{errors.url.message}</p>
                        )}
                        <p className="text-xs text-gray-500">
                            Must use HTTPS for security
                        </p>
                    </div>

                    {/* Events */}
                    <div className="space-y-3">
                        <Label>Events to Subscribe *</Label>
                        <div className="space-y-2 max-h-64 overflow-y-auto">
                            {AVAILABLE_EVENTS.map((event) => (
                                <div key={event.value} className="flex items-start gap-3 p-3 border rounded-lg">
                                    <Checkbox
                                        id={`event-${event.value}`}
                                        checked={selectedEvents.includes(event.value)}
                                        onCheckedChange={(checked) => {
                                            if (checked) {
                                                setValue('events', [...selectedEvents, event.value]);
                                            } else {
                                                setValue('events', selectedEvents.filter(e => e !== event.value));
                                            }
                                        }}
                                        disabled={createWebhookMutation.isPending}
                                    />
                                    <div className="flex-1">
                                        <Label htmlFor={`event-${event.value}`} className="font-medium cursor-pointer">
                                            {event.label}
                                        </Label>
                                        <p className="text-xs text-gray-500 mt-0.5">{event.description}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                        {errors.events && (
                            <p className="text-sm text-red-600">{errors.events.message}</p>
                        )}
                    </div>

                    {/* Security Warning */}
                    <Alert className="bg-orange-50 border-orange-200">
                        <AlertTriangle className="h-4 w-4 text-orange-600" />
                        <AlertDescription className="text-orange-800 text-sm">
                            <strong>Security:</strong> Webhooks will receive sensitive event data.
                            Ensure your endpoint is secure and verifies signatures.
                        </AlertDescription>
                    </Alert>

                    <DialogFooter>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleClose}
                            disabled={createWebhookMutation.isPending}
                        >
                            Cancel
                        </Button>
                        <Button type="submit" disabled={createWebhookMutation.isPending}>
                            {createWebhookMutation.isPending ? 'Creating...' : 'Create Webhook'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
