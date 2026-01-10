/**
 * Create SCIM Token Dialog
 * 
 * AUTHORITY MODEL:
 * Who: TENANT users OR SYSTEM users (tenant selected)
 * Scope: Tenant-scoped SCIM token
 * Permission: scim:tokens:create
 * 
 * SECURITY GUARDRAILS:
 * - One-time secret display only
 * - No token re-display
 * - Security warnings at creation
 * - Confirmation before closing without copying
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
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Copy, CheckCircle2, AlertTriangle, Shield } from 'lucide-react';

const createSCIMTokenSchema = z.object({
    name: z.string().min(1, 'Token name is required'),
});

type CreateSCIMTokenFormData = z.infer<typeof createSCIMTokenSchema>;

interface CreateSCIMTokenDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    tenantId: string;
}

export function CreateSCIMTokenDialog({ open, onOpenChange, tenantId }: CreateSCIMTokenDialogProps) {
    const queryClient = useQueryClient();
    const [error, setError] = useState<string | null>(null);
    const [createdToken, setCreatedToken] = useState<{ token: string; name: string } | null>(null);
    const [tokenCopied, setTokenCopied] = useState(false);

    const {
        register,
        handleSubmit,
        formState: { errors },
        reset,
    } = useForm<CreateSCIMTokenFormData>({
        resolver: zodResolver(createSCIMTokenSchema),
    });

    const createTokenMutation = useMutation({
        mutationFn: async (data: CreateSCIMTokenFormData) => {
            // TODO: Implement API call
            // return scimApi.createToken(tenantId, data);

            // Placeholder response
            return {
                token: 'scim_' + Math.random().toString(36).substr(2, 48),
                name: data.name,
            };
        },
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['scim-tokens', tenantId] });
            setCreatedToken(data);
            setError(null);
        },
        onError: (err: any) => {
            setError(err.message || 'Failed to create SCIM token');
        },
    });

    const onSubmit = (data: CreateSCIMTokenFormData) => {
        setError(null);
        createTokenMutation.mutate(data);
    };

    const handleClose = () => {
        if (createdToken && !tokenCopied) {
            if (!confirm('You have not copied the SCIM token. It will not be shown again. Are you sure you want to close?')) {
                return;
            }
        }
        onOpenChange(false);
        reset();
        setCreatedToken(null);
        setTokenCopied(false);
        setError(null);
    };

    const copyToken = () => {
        if (createdToken) {
            navigator.clipboard.writeText(createdToken.token);
            setTokenCopied(true);
        }
    };

    // Show success screen after creation
    if (createdToken) {
        return (
            <Dialog open={open} onOpenChange={handleClose}>
                <DialogContent className="max-w-md">
                    <DialogHeader>
                        <DialogTitle>SCIM Token Created</DialogTitle>
                        <DialogDescription>
                            Save this token securely. It will not be shown again.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="space-y-4">
                        <Alert className="bg-green-50 border-green-200">
                            <CheckCircle2 className="h-4 w-4 text-green-600" />
                            <AlertDescription className="text-green-800">
                                Token "{createdToken.name}" created successfully
                            </AlertDescription>
                        </Alert>

                        <div className="space-y-2">
                            <Label>SCIM Token</Label>
                            <div className="flex gap-2">
                                <Input
                                    value={createdToken.token}
                                    readOnly
                                    className="font-mono text-sm"
                                    type="password"
                                />
                                <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={copyToken}
                                    className={tokenCopied ? 'bg-green-50' : ''}
                                >
                                    {tokenCopied ? (
                                        <CheckCircle2 className="h-4 w-4 text-green-600" />
                                    ) : (
                                        <Copy className="h-4 w-4" />
                                    )}
                                </Button>
                            </div>
                            <p className="text-xs text-red-600">
                                ⚠️ This token will not be shown again. Save it securely.
                            </p>
                        </div>

                        <Alert className="bg-orange-50 border-orange-200">
                            <Shield className="h-4 w-4 text-orange-600" />
                            <AlertDescription className="text-orange-800 text-sm">
                                <strong>Security:</strong> This token grants full SCIM provisioning access.
                                Treat it as a root credential and store it in a secure location (e.g., secrets manager).
                            </AlertDescription>
                        </Alert>
                    </div>

                    <DialogFooter>
                        <Button onClick={handleClose} variant={tokenCopied ? 'default' : 'outline'}>
                            {tokenCopied ? 'Done' : 'I have saved the token'}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        );
    }

    return (
        <Dialog open={open} onOpenChange={handleClose}>
            <DialogContent className="max-w-md">
                <DialogHeader>
                    <DialogTitle>Create SCIM Token</DialogTitle>
                    <DialogDescription>
                        Create a new authentication token for SCIM provisioning
                    </DialogDescription>
                </DialogHeader>

                <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                    {error && (
                        <Alert variant="destructive">
                            <AlertDescription>{error}</AlertDescription>
                        </Alert>
                    )}

                    <div className="space-y-2">
                        <Label htmlFor="name">Token Name *</Label>
                        <Input
                            id="name"
                            {...register('name')}
                            placeholder="Production IdP"
                            disabled={createTokenMutation.isPending}
                        />
                        {errors.name && (
                            <p className="text-sm text-red-600">{errors.name.message}</p>
                        )}
                        <p className="text-xs text-gray-500">
                            A descriptive name to identify this token (e.g., "Okta Production", "Azure AD")
                        </p>
                    </div>

                    <Alert className="bg-orange-50 border-orange-200">
                        <AlertTriangle className="h-4 w-4 text-orange-600" />
                        <AlertDescription className="text-orange-800 text-sm">
                            <strong>Security:</strong> The token will be shown only once.
                            You will not be able to retrieve it after closing this dialog.
                        </AlertDescription>
                    </Alert>

                    <DialogFooter>
                        <Button
                            type="button"
                            variant="outline"
                            onClick={handleClose}
                            disabled={createTokenMutation.isPending}
                        >
                            Cancel
                        </Button>
                        <Button type="submit" disabled={createTokenMutation.isPending}>
                            {createTokenMutation.isPending ? 'Creating...' : 'Create Token'}
                        </Button>
                    </DialogFooter>
                </form>
            </DialogContent>
        </Dialog>
    );
}
