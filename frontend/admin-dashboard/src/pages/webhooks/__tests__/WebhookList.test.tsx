import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { WebhookList } from '../WebhookList';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { webhookApi } from '@/services/api';

// Mock dependencies
vi.mock('@/services/api', () => ({
    webhookApi: {
        list: vi.fn(),
        create: vi.fn(),
        delete: vi.fn(),
    },
}));

const mockContext = {
    principalType: 'SYSTEM',
    selectedTenantId: 'tenant-123',
    homeTenantId: null,
};

vi.mock('@/contexts/PrincipalContext', () => ({
    usePrincipalContext: () => mockContext,
}));

// Mock PermissionGate to render children
vi.mock('@/components/PermissionGate', () => ({
    PermissionGate: ({ children }: any) => <div>{children}</div>,
}));

// Setup QueryClient
const createTestQueryClient = () => new QueryClient({
    defaultOptions: {
        queries: { retry: false },
    },
});

describe('WebhookList', () => {
    let queryClient: QueryClient;

    beforeEach(() => {
        queryClient = createTestQueryClient();
        vi.clearAllMocks();
    });

    it('renders list of webhooks', async () => {
        const mockWebhooks = [
            { id: '1', name: 'Test Webhook', url: 'https://example.com', enabled: true, events: ['user.created'], created_at: new Date().toISOString() },
        ];
        (webhookApi.list as any).mockResolvedValue(mockWebhooks);

        render(
            <QueryClientProvider client={queryClient}>
                <WebhookList />
            </QueryClientProvider>
        );

        await waitFor(() => {
            expect(screen.getByText('Test Webhook')).toBeInTheDocument();
            expect(screen.getByText('https://example.com')).toBeInTheDocument();
        });
    });

});
});
