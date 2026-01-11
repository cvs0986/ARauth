import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { CreateWebhookDialog } from '../CreateWebhookDialog';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { webhookApi } from '@/services/api';

vi.mock('@/services/api', () => ({
    webhookApi: {
        create: vi.fn(),
    },
}));

const createTestQueryClient = () => new QueryClient({
    defaultOptions: {
        queries: { retry: false },
    },
});

describe('CreateWebhookDialog', () => {
    let queryClient: QueryClient;
    const onOpenChange = vi.fn();

    beforeEach(() => {
        queryClient = createTestQueryClient();
        vi.clearAllMocks();
    });

    it('renders form elements', () => {
        render(
            <QueryClientProvider client={queryClient}>
                <CreateWebhookDialog open={true} onOpenChange={onOpenChange} scope="tenant" tenantId="tenant-1" />
            </QueryClientProvider>
        );

        expect(screen.getByText('Create Webhook')).toBeInTheDocument();
    });

    it('submits form and displays secret on success', async () => {
        const mockCreatedWebhook = {
            id: '1',
            name: 'New Webhook',
            secret: 'whsec_test123',
            // ... other fields
        };
        (webhookApi.create as any).mockResolvedValue(mockCreatedWebhook);

        render(
            <QueryClientProvider client={queryClient}>
                <CreateWebhookDialog open={true} onOpenChange={onOpenChange} scope="tenant" tenantId="tenant-1" />
            </QueryClientProvider>
        );

        // Fill form
        fireEvent.change(screen.getByLabelText(/Name/i), { target: { value: 'New Webhook' } });
        fireEvent.change(screen.getByLabelText(/Endpoint URL/i), { target: { value: 'https://example.com' } });

        // Select events (User Created) - Assuming label matches
        const userCreatedParams = screen.getByLabelText(/User Created/i);
        fireEvent.click(userCreatedParams);

        // Submit
        const createButton = screen.getByRole('button', { name: 'Create Webhook' });
        fireEvent.click(createButton);

        await waitFor(() => {
            expect(webhookApi.create).toHaveBeenCalledWith(expect.objectContaining({
                name: 'New Webhook',
                url: 'https://example.com',
                events: expect.arrayContaining(['user.created']),
            }));
        });

        // Assert Secret Display
        await waitFor(() => {
            expect(screen.getByText('Webhook Created Successfully')).toBeInTheDocument();
            expect(screen.getByDisplayValue('whsec_test123')).toBeInTheDocument();
        });
    });
});
