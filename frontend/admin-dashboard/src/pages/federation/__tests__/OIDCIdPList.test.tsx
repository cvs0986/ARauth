import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { OIDCIdPList } from '../OIDCIdPList';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { federationApi } from '@/services/api';

vi.mock('@/services/api', () => ({
    federationApi: {
        list: vi.fn(),
        verify: vi.fn(),
    },
}));

vi.mock('@/contexts/PrincipalContext', () => ({
    usePrincipalContext: () => ({
        principalType: 'SYSTEM',
        selectedTenantId: 'tenant-123',
    }),
}));

vi.mock('@/components/PermissionGate', () => ({
    PermissionGate: ({ children }: any) => <div>{children}</div>,
}));

// Mock window.alert
const mockAlert = vi.spyOn(window, 'alert').mockImplementation(() => { });

const createTestQueryClient = () => new QueryClient({
    defaultOptions: {
        queries: { retry: false },
    },
});

describe('OIDCIdPList', () => {
    let queryClient: QueryClient;

    beforeEach(() => {
        queryClient = createTestQueryClient();
        vi.clearAllMocks();
    });

    it('renders list of oidc providers', async () => {
        const mockIdps = [{
            id: '1',
            name: 'Google',
            type: 'oidc',
            enabled: true,
            configuration: { issuer_url: 'https://accounts.google.com', client_id: '123' },
            created_at: new Date().toISOString()
        }];
        (federationApi.list as any).mockResolvedValue(mockIdps);

        render(
            <QueryClientProvider client={queryClient}>
                <OIDCIdPList />
            </QueryClientProvider>
        );

        await waitFor(() => {
            expect(screen.getByText('Google')).toBeInTheDocument();
            expect(screen.getByText('https://accounts.google.com')).toBeInTheDocument();
        });
    });

    it('calls verify api when test connection is clicked', async () => {
        const mockIdps = [{
            id: '1',
            name: 'Google',
            type: 'oidc',
            enabled: true,
            configuration: { issuer_url: 'https://accounts.google.com' },
            created_at: new Date().toISOString()
        }];
        (federationApi.list as any).mockResolvedValue(mockIdps);
        (federationApi.verify as any).mockResolvedValue({ success: true, message: 'Connected' });

        render(
            <QueryClientProvider client={queryClient}>
                <OIDCIdPList />
            </QueryClientProvider>
        );

        await waitFor(() => expect(screen.getByText('Google')).toBeInTheDocument());

        const verifyButton = screen.getByRole('button', { name: /Test Connection/i });
        fireEvent.click(verifyButton);

        await waitFor(() => {
            expect(federationApi.verify).toHaveBeenCalledWith('1');
            expect(mockAlert).toHaveBeenCalledWith(expect.stringContaining('Connection Successful'));
        });
    });
});
