import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { AuditLogList } from '../AuditLogList';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { auditApi } from '@/services/api';

vi.mock('@/services/api', () => ({
    auditApi: {
        list: vi.fn(),
        export: vi.fn(),
    },
}));

vi.mock('@/contexts/PrincipalContext', () => ({
    usePrincipalContext: () => ({
        principalType: 'SYSTEM',
        selectedTenantId: 'tenant-123',
    }),
})); // Mock PrincipalContext to ensure effectiveTenantId is set

// Mock PermissionGate to render children
vi.mock('@/components/PermissionGate', () => ({
    PermissionGate: ({ children }: any) => <div>{children}</div>,
}));

const createTestQueryClient = () => new QueryClient({
    defaultOptions: {
        queries: { retry: false },
    },
});

describe('AuditLogList', () => {
    let queryClient: QueryClient;

    beforeEach(() => {
        queryClient = createTestQueryClient();
        vi.clearAllMocks();
    });

    it('renders list of logs', async () => {
        const mockLogs = [
            { id: '1', event_type: 'user.login', timestamp: new Date().toISOString(), actor: { username: 'user1' }, result: 'success' }
        ];
        (auditApi.list as any).mockResolvedValue(mockLogs);

        render(
            <QueryClientProvider client={queryClient}>
                <AuditLogList />
            </QueryClientProvider>
        );

        await waitFor(() => {
            expect(screen.getByText('user.login')).toBeInTheDocument();
            expect(screen.getByText('user1')).toBeInTheDocument();
        });
    });

    it('export calls api', async () => {
        const mockLogs: any[] = [];
        (auditApi.list as any).mockResolvedValue(mockLogs);
        (auditApi.export as any).mockResolvedValue(undefined);

        render(
            <QueryClientProvider client={queryClient}>
                <AuditLogList />
            </QueryClientProvider>
        );

        await waitFor(() => expect(screen.getByText('Audit Logs')).toBeInTheDocument());

        const exportButton = screen.getByText('Export'); // Assuming button text is "Export"
        fireEvent.click(exportButton);

        await waitFor(() => {
            expect(auditApi.export).toHaveBeenCalled();
        });
    });
});
