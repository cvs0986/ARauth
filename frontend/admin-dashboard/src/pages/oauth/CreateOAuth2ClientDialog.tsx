/**
 * Create OAuth2 Client Dialog
 * 
 * AUTHORITY MODEL:
 * Who: TENANT users OR SYSTEM users (tenant selected)
 * Scope: Tenant-scoped OAuth2 client
 * Permission: oauth:clients:create
 * 
 * GUARDRAILS:
 * - No default scopes
 * - No auto-grant types
 * - Explicit redirect URI validation
 * - Secret shown once on creation
 */

import { useState } from 'react';
import { useForm, useFieldArray } from 'react-hook-form';
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
import { Plus, X, Key, AlertTriangle, Copy, CheckCircle2 } from 'lucide-react';

const createOAuth2ClientSchema = z.object({
    client_name: z.string().min(1, 'Client name is required'),
    grant_types: z.array(z.string()).min(1, 'At least one grant type is required'),
    redirect_uris: z.array(z.string().url('Must be a valid URL')).min(1, 'At least one redirect URI is required'),
    scopes: z.array(z.string()).optional(),
});

type CreateOAuth2ClientFormData = z.infer<typeof createOAuth2ClientSchema>;

interface CreateOAuth2ClientDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    tenantId: string;
}

const AVAILABLE_GRANT_TYPES = [
    { value: 'authorization_code', label: 'Authorization Code', description: 'Standard OAuth2 flow with PKCE' },
    { value: 'client_credentials', label: 'Client Credentials', description: 'Machine-to-machine' },
    { value: 'refresh_token', label: 'Refresh Token', description: 'Token refresh capability' },
];

const AVAILABLE_SCOPES = [
    { value: 'openid', label: 'OpenID', description: 'OIDC authentication' },
    { value: 'profile', label: 'Profile', description: 'User profile information' },
    { value: 'email', label: 'Email', description: 'User email address' },
    { value: 'offline_access', label: 'Offline Access', description: 'Refresh token issuance' },
];

export function CreateOAuth2ClientDialog({ open, onOpenChange, tenantId }: CreateOAuth2ClientDialogProps) {
    const queryClient = useQueryClient();
    const [error, setError] = useState<string | null>(null);
    const [createdClient, setCreatedClient] = useState<{ client_id: string; client_secret: string } | null>(null);
    const [secretCopied, setSecretCopied] = useState(false);

    const {
        register,
        handleSubmit,
        formState: { errors },
        reset,
        control,
        watch,
        setValue,
    } = useForm<CreateOAuth2ClientFormData>({
        resolver: zodResolver(createOAuth2ClientSchema),
        defaultValues: {
            client_name: '',
            grant_types: [],
            redirect_uris: [''],
            scopes: [],
        },
    });

    const { fields: redirectUriFields, append: appendRedirectUri, remove: removeRedirectUri } = useFieldArray({
        control,
        name: 'redirect_uris',
    });

    const selectedGrantTypes = watch('grant_types') || [];
    const selectedScopes = watch('scopes') || [];

    const createClientMutation = useMutation({
        mutationFn: async (data: CreateOAuth2ClientFormData) => {
            // TODO: Implement API call
            // return oauthApi.createClient(tenantId, data);

            // Placeholder response
            return {
                client_id: 'oauth_' + Math.random().toString(36).substr(2, 9),
                client_secret: 'secret_' + Math.random().toString(36).substr(2, 32),
                ...data,
            };
        },
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['oauth-clients', tenantId] });
            setCreatedClient({ client_id: data.client_id, client_secret: data.client_secret });
            setError(null);
        },
        onError: (err: any) => {
            setError(err.message || 'Failed to create OAuth2 client');
        },
    });

    const onSubmit = (data: CreateOAuth2ClientFormData) => {
        setError(null);
        createClientMutation.mutate(data);
    };

    const handleClose = () => {
        if (createdClient && !secretCopied) {
            if (!confirm('You have not copied the client secret. It will not be shown again. Are you sure you want to close?')) {
                return;
            }
        }
        onOpenChange(false);
        reset();
        setCreatedClient(null);
        setSecretCopied(false);
        setError(null);
    };

    const copySecret = () => {
        if (createdClient) {
            navigator.clipboard.writeText(createdClient.client_secret);
            setSecretCopied(true);
        }
    };

    // Show success screen after creation
    if (createdClient) {
        return (
            <Dialog open={open} onOpenChange={handleClose}>
                <DialogContent className="max-w-md">
                    <DialogHeader>
                        <DialogTitle>OAuth2 Client Created</DialogTitle>
                        <DialogDescription>
                            Save these credentials securely. The client secret will not be shown again.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="space-y-4">
                        <Alert className="bg-green-50 border-green-200">
                            <CheckCircle2 className="h-4 w-4 text-green-600" />
                            <AlertDescription className="text-green-800">
                                Client created successfully
                            </AlertDescription>
                        </Alert>

                        <div className="space-y-2">
                            <Label>Client ID</Label>
                            <div className="flex gap-2">
                                <Input value={createdClient.client_id} readOnly className="font-mono text-sm" />
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => navigator.clipboard.writeText(createdClient.client_id)}
                                >
                                    <Copy className="h-4 w-4" />
                                </Button>
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label>Client Secret</Label>
                            <div className="flex gap-2">
                                <Input value={createdClient.client_secret} readOnly className="font-mono text-sm" />
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={copySecret}
                                    className={secretCopied ? 'bg-green-50' : ''}
                                >
                                    {secretCopied ? <CheckCircle2 className="h-4 w-4 text-green-600" /> : <Copy className="h-4 w-4" />}
                                </Button>
                            </div>
                            <p className="text-xs text-red-600">
                                ⚠️ This secret will not be shown again. Save it securely.
                            </p>
                        </div>
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
                    <DialogTitle>Create OAuth2 Client</DialogTitle>
                    <DialogDescription>
                        Create a new OAuth2/OIDC client application for this tenant
                    </DialogDescription>
                </DialogHeader>

                <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
                    {error && (
                        <Alert variant="destructive">
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    {/* Client Name */}
                    <div className="space-y-2">
                        <Label htmlFor="client_name">Client Name *</Label>
                        <Input
                            id="client_name"
                            {...register('client_name')}
                            placeholder="My Application"
                            disabled={createClientMutation.isPending}
                        />
                        {errors.client_name && (
                            <p className="text-sm text-red-600">{errors.client_name.message}</p>
                        )}
                    </div>

                    {/* Grant Types */}
                    <div className="space-y-3">
                        <Label>Grant Types *</Label>
                        <div className="space-y-2">
                            {AVAILABLE_GRANT_TYPES.map((grantType) => (
                                <div key={grantType.value} className="flex items-start gap-3 p-3 border rounded-lg">
                                    <Checkbox
                                        id={`grant-${grantType.value}`}
                                        checked={selectedGrantTypes.includes(grantType.value)}
                                        onCheckedChange={(checked) => {
                                            if (checked) {
                                                setValue('grant_types', [...selectedGrantTypes, grantType.value]);
                                            } else {
                                                setValue('grant_types', selectedGrantTypes.filter(g => g !== grantType.value));
                                            }
                                        }}
                                        disabled={createClientMutation.isPending}
                                    />
                                    <div className="flex-1">
                                        <Label htmlFor={`grant-${grantType.value}`} className="font-medium cursor-pointer">
                                            {grantType.label}
                                        </Label>
                                        <p className="text-xs text-gray-500 mt-0.5">{grantType.description}</p>
                                    </div>
                                </div>
                            ))}
                        </div>
                        {errors.grant_types && (
                            <p className="text-sm text-red-600">{errors.grant_types.message}</p>
                        )}
                    </div>

                    {/* Redirect URIs */}
                    <div className="space-y-3">
                        <Label>Redirect URIs *</Label>
                        {redirectUriFields.map((field, index) => (
                            <div key={field.id} className="flex gap-2">
                                <Input
                                    {...register(`redirect_uris.${index}` as const)}
                                    placeholder="https://example.com/callback"
                                    disabled={createClientMutation.isPending}
                                />
                                {redirectUriFields.length > 1 && (
                                    <Button
                                        type="button"
                                        variant="outline"
                                        size="sm"
                                        onClick={() => removeRedirectUri(index)}
                                        disabled={createClientMutation.isPending}
                                    >
                                        <X className="h-4 w-4" />
                                    </Button>
                                )}
                            </div>
                        ))}
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => appendRedirectUri('')}
                            disabled={createClientMutation.isPending}
                        >
                            <Plus className="h-4 w-4 mr-2" />
                            Add Redirect URI
                        </Button>
                        {errors.redirect_uris && (
                            <p className="text-sm text-red-600">
                                {errors.redirect_uris.message || errors.redirect_uris.root?.message}
                            </p>
                        )}
                    </div>

                    {/* Scopes */}
                    <div className="space-y-3">
                        <Label>Scopes (Optional)</Label>
                        <div className="flex flex-wrap gap-2">
                            {AVAILABLE_SCOPES.map((scope) => (
                                <Badge
                                    key={scope.value}
                                    variant={selectedScopes.includes(scope.value) ? 'default' : 'outline'}
                                    className="cursor-pointer"
                                    onClick={() => {
                                        if (selectedScopes.includes(scope.value)) {
                                            setValue('scopes', selectedScopes.filter(s => s !== scope.value));
                                        } else {
                                            setValue('scopes', [...selectedScopes, scope.value]);
                                        }
                                    }}
                                >
                                    {scope.label}
                                </Badge>
                            ))}
                        </div>
                        <p className="text-xs text-gray-500">
                            Select the scopes this client can request. Leave empty to allow all scopes.
                        </p>
                    </div>

                    {/* Security Notice */}
                    <Alert className="bg-orange-50 border-orange-200">
                        <AlertTriangle className="h-4 w-4 text-orange-600" />
                        <AlertDescription className="text-orange-800 text-sm">
                            <strong>Security:</strong> The client secret will be shown only once. Store it securely.
                        </AlertDescription>
                    </Alert>

                    <DialogFooter>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleClose}
                            disabled={createClientMutation.isPending}
                        >
                            Cancel
                        </Button>
                        <Button type="submit" disabled={createClientMutation.isPending}>
                            {createClientMutation.isPending ? 'Creating...' : 'Create Client'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
