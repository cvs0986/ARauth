import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { OAuth2ClientList } from '../OAuth2ClientList';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { oauthClientApi } from '@/services/api';

vi.mock('@/services/api', () => ({
    oauthClientApi: {
        list: vi.fn(),
        rotateSecret: vi.fn(),
        delete: vi.fn(),
    },
}));

vi.mock('@/contexts/PrincipalContext', () => ({
    usePrincipalContext: () => ({
        principalType: 'SYSTEM',
        selectedTenantId: 'tenant-123',
        consoleMode: 'SYSTEM',
    }),
}));

vi.mock('@/components/PermissionGate', () => ({
    PermissionGate: ({ children }: any) => <div>{children}</div>,
}));

// Mock window interactions
const mockPrompt = vi.spyOn(window, 'prompt').mockImplementation(() => 'ok');
const mockConfirm = vi.spyOn(window, 'confirm').mockReturnValue(true);

const createTestQueryClient = () => new QueryClient({
    defaultOptions: {
        queries: { retry: false },
    },
});

describe('OAuth2ClientList', () => {
    let queryClient: QueryClient;

    beforeEach(() => {
        queryClient = createTestQueryClient();
        vi.clearAllMocks();
    });

    it('renders list of clients', async () => {
        const mockClients = [{
            id: '1',
            name: 'App1',
            client_name: 'App1',
            client_id: 'client_123',
            grant_types: ['authorization_code'],
            redirect_uris: ['http://localhost/cb'],
            created_at: new Date().toISOString()
        }];
        (oauthClientApi.list as any).mockResolvedValue(mockClients);

        render(
            <QueryClientProvider client={queryClient}>
                <OAuth2ClientList />
            </QueryClientProvider>
        );

        await waitFor(() => {
            expect(screen.getByText('App1')).toBeInTheDocument();
            expect(screen.getByText('client_123')).toBeInTheDocument();
        });
    });

    it('rotates secret when confirmed', async () => {
        const mockClients = [{
            id: '1',
            name: 'App1',
            client_name: 'App1',
            client_id: 'client_123',
            grant_types: [],
            redirect_uris: [],
            created_at: new Date().toISOString()
        }];
        (oauthClientApi.list as any).mockResolvedValue(mockClients);
        (oauthClientApi.rotateSecret as any).mockResolvedValue({ client_secret: 'new_secret' });

        render(
            <QueryClientProvider client={queryClient}>
                <OAuth2ClientList />
            </QueryClientProvider>
        );

        await waitFor(() => expect(screen.getByText('App1')).toBeInTheDocument());

        const rotateButton = screen.getByTitle('Rotate Secret');
        fireEvent.click(rotateButton);

        expect(mockConfirm).toHaveBeenCalled();
        expect(oauthClientApi.rotateSecret).toHaveBeenCalledWith('1');

        await waitFor(() => {
            expect(mockPrompt).toHaveBeenCalledWith(expect.stringContaining('Copy the new secret'), 'new_secret');
        });
    });
});
