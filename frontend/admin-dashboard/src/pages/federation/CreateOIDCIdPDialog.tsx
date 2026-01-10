/**
 * Create OIDC Identity Provider Dialog
 * 
 * AUTHORITY MODEL:
 * Who: TENANT users OR SYSTEM users (tenant selected)
 * Scope: Tenant-scoped OIDC IdP
 * Permission: federation:idp:create
 * 
 * SECURITY:
 * - One-time client secret display
 * - Explicit attribute mapping (no auto-mapping)
 * - Test connection before enabling
 * - Clear warnings for login impact
 * 
 * UI CONTRACT MODE:
 * - createIdP throws APINotConnectedError
 * - testConnection throws APINotConnectedError
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
import { Textarea } from '@/components/ui/textarea';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Badge } from '@/components/ui/badge';
import { Checkbox } from '@/components/ui/checkbox';
import { AlertTriangle, Info, TestTube } from 'lucide-react';
import { APINotConnectedError, getAPINotConnectedMessage, isAPINotConnected } from '@/lib/errors';

const createOIDCIdPSchema = z.object({
    name: z.string().min(1, 'Provider name is required'),
    issuer_url: z.string().url('Must be a valid URL'),
    client_id: z.string().min(1, 'Client ID is required'),
    client_secret: z.string().min(1, 'Client secret is required'),
    scopes: z.string().min(1, 'At least one scope is required'),
    attribute_mapping: z.object({
        email: z.string().min(1, 'Email mapping is required'),
        name: z.string().optional(),
        given_name: z.string().optional(),
        family_name: z.string().optional(),
    }),
});

type CreateOIDCIdPFormData = z.infer<typeof createOIDCIdPSchema>;

interface CreateOIDCIdPDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    tenantId: string;
}

export function CreateOIDCIdPDialog({ open, onOpenChange, tenantId }: CreateOIDCIdPDialogProps) {
    const queryClient = useQueryClient();
    const [error, setError] = useState<string | null>(null);
    const [testResult, setTestResult] = useState<'success' | 'error' | null>(null);

    const {
        register,
        handleSubmit,
        formState: { errors },
        reset,
        watch,
    } = useForm<CreateOIDCIdPFormData>({
        resolver: zodResolver(createOIDCIdPSchema),
        defaultValues: {
            scopes: 'openid profile email',
            attribute_mapping: {
                email: 'email',
                name: 'name',
                given_name: 'given_name',
                family_name: 'family_name',
            },
        },
    });

    const createIdPMutation = useMutation({
        mutationFn: async (data: CreateOIDCIdPFormData) => {
            // UI CONTRACT MODE: Throw APINotConnectedError
            throw new APINotConnectedError('federation.oidc.create');
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['oidc-idps', tenantId] });
            reset();
            setError(null);
            onOpenChange(false);
        },
        onError: (err: any) => {
            setError(getAPINotConnectedMessage(err));
        },
    });

    const testConnectionMutation = useMutation({
        mutationFn: async (data: CreateOIDCIdPFormData) => {
            // UI CONTRACT MODE: Throw APINotConnectedError
            throw new APINotConnectedError('federation.oidc.testConnection');
        },
        onSuccess: () => {
            setTestResult('success');
        },
        onError: (err: any) => {
            setTestResult('error');
            setError(getAPINotConnectedMessage(err));
        },
    });

    const onSubmit = (data: CreateOIDCIdPFormData) => {
        setError(null);
        createIdPMutation.mutate(data);
    };

    const handleTestConnection = () => {
        const formData = watch();
        setError(null);
        setTestResult(null);
        testConnectionMutation.mutate(formData as CreateOIDCIdPFormData);
    };

    const handleClose = () => {
        onOpenChange(false);
        reset();
        setError(null);
        setTestResult(null);
    };

    return (
        <Dialog open={open} onOpenChange={handleClose}>
            <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Add OIDC Identity Provider</DialogTitle>
                    <DialogDescription>
                        Configure an external OpenID Connect provider for federated authentication
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

                    {/* Provider Name */}
                    <div className="space-y-2">
                        <Label htmlFor="name">Provider Name *</Label>
                        <Input
                            id="name"
                            {...register('name')}
                            placeholder="Google Workspace"
                            disabled={createIdPMutation.isPending}
                        />
                        {errors.name && (
                            <p className="text-sm text-red-600">{errors.name.message}</p>
                        )}
                    </div>

                    {/* Issuer URL */}
                    <div className="space-y-2">
                        <Label htmlFor="issuer_url">Issuer URL *</Label>
                        <Input
                            id="issuer_url"
                            {...register('issuer_url')}
                            placeholder="https://accounts.google.com"
                            disabled={createIdPMutation.isPending}
                        />
                        {errors.issuer_url && (
                            <p className="text-sm text-red-600">{errors.issuer_url.message}</p>
                        )}
                        <p className="text-xs text-gray-500">
                            The OIDC issuer URL (e.g., https://accounts.google.com)
                        </p>
                    </div>

                    {/* Client ID */}
                    <div className="space-y-2">
                        <Label htmlFor="client_id">Client ID *</Label>
                        <Input
                            id="client_id"
                            {...register('client_id')}
                            placeholder="your-client-id.apps.googleusercontent.com"
                            disabled={createIdPMutation.isPending}
                            className="font-mono text-sm"
                        />
                        {errors.client_id && (
                            <p className="text-sm text-red-600">{errors.client_id.message}</p>
                        )}
                    </div>

                    {/* Client Secret */}
                    <div className="space-y-2">
                        <Label htmlFor="client_secret">Client Secret *</Label>
                        <Input
                            id="client_secret"
                            type="password"
                            {...register('client_secret')}
                            placeholder="Enter client secret"
                            disabled={createIdPMutation.isPending}
                            className="font-mono text-sm"
                        />
                        {errors.client_secret && (
                            <p className="text-sm text-red-600">{errors.client_secret.message}</p>
                        )}
                    </div>

                    {/* Scopes */}
                    <div className="space-y-2">
                        <Label htmlFor="scopes">Scopes *</Label>
                        <Input
                            id="scopes"
                            {...register('scopes')}
                            placeholder="openid profile email"
                            disabled={createIdPMutation.isPending}
                        />
                        {errors.scopes && (
                            <p className="text-sm text-red-600">{errors.scopes.message}</p>
                        )}
                        <p className="text-xs text-gray-500">
                            Space-separated list of OIDC scopes to request
                        </p>
                    </div>

                    {/* Attribute Mapping */}
                    <div className="space-y-4">
                        <div className="flex items-center justify-between">
                            <Label>Attribute Mapping *</Label>
                            <Badge variant="outline" className="text-xs">
                                Explicit Mapping Required
                            </Badge>
                        </div>
                        <Alert className="bg-blue-50 border-blue-200">
                            <Info className="h-4 w-4 text-blue-600" />
                            <AlertDescription className="text-blue-800 text-sm">
                                Map OIDC claims to ARauth user attributes. Email mapping is required.
                            </AlertDescription>
                        </Alert>

                        <div className="grid grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="email_mapping">Email Claim *</Label>
                                <Input
                                    id="email_mapping"
                                    {...register('attribute_mapping.email')}
                                    placeholder="email"
                                    disabled={createIdPMutation.isPending}
                                    className="font-mono text-sm"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="name_mapping">Name Claim</Label>
                                <Input
                                    id="name_mapping"
                                    {...register('attribute_mapping.name')}
                                    placeholder="name"
                                    disabled={createIdPMutation.isPending}
                                    className="font-mono text-sm"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="given_name_mapping">Given Name Claim</Label>
                                <Input
                                    id="given_name_mapping"
                                    {...register('attribute_mapping.given_name')}
                                    placeholder="given_name"
                                    disabled={createIdPMutation.isPending}
                                    className="font-mono text-sm"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="family_name_mapping">Family Name Claim</Label>
                                <Input
                                    id="family_name_mapping"
                                    {...register('attribute_mapping.family_name')}
                                    placeholder="family_name"
                                    disabled={createIdPMutation.isPending}
                                    className="font-mono text-sm"
                                />
                            </div>
                        </div>
                    </div>

                    {/* Test Connection */}
                    <div className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                        <div>
                            <Label className="text-sm font-medium">Test Connection</Label>
                            <p className="text-xs text-gray-500 mt-0.5">
                                Verify the OIDC configuration before saving
                            </p>
                        </div>
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={handleTestConnection}
                            disabled={testConnectionMutation.isPending}
                        >
                            <TestTube className="h-4 w-4 mr-2" />
                            {testConnectionMutation.isPending ? 'Testing...' : 'Test'}
                        </Button>
                    </div>

                    {testResult && (
                        <Alert variant={testResult === 'success' ? 'default' : 'destructive'}>
                            <AlertDescription>
                                {testResult === 'success'
                                    ? 'Connection test successful'
                                    : 'Connection test failed'}
                            </AlertDescription>
                        </Alert>
                    )}

                    {/* Security Warning */}
                    <Alert className="bg-orange-50 border-orange-200">
                        <AlertTriangle className="h-4 w-4 text-orange-600" />
                        <AlertDescription className="text-orange-800 text-sm">
                            <strong>Security:</strong> This provider will be created in disabled state.
                            Test thoroughly before enabling for user authentication.
                        </AlertDescription>
                    </Alert>

                    <DialogFooter>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleClose}
                            disabled={createIdPMutation.isPending}
                        >
                            Cancel
                        </Button>
                        <Button type="submit" disabled={createIdPMutation.isPending}>
                            {createIdPMutation.isPending ? 'Creating...' : 'Create Provider'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
