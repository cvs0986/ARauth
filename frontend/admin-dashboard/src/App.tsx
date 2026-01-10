/**
 * Main App Component
 */

import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { PrincipalProvider } from './contexts/PrincipalContext';
import { ProtectedRoute } from './components/ProtectedRoute';
import { Layout } from './components/layout/Layout';
import { Login } from './pages/Login';
import { NoAccess } from './pages/NoAccess';
import { Dashboard } from './pages/Dashboard';
import { Settings } from './pages/Settings';
import { AuditLogs } from './pages/AuditLogs';
import { TenantList } from './pages/tenants/TenantList';
import { TenantDetail } from './pages/tenants/TenantDetail';
import { UserList } from './pages/users/UserList';
import { UserDetail } from './pages/users/UserDetail';
import { RoleList } from './pages/roles/RoleList';
import { PermissionList } from './pages/permissions/PermissionList';
import { MFA } from './pages/MFA';
import { SystemCapabilityList } from './pages/capabilities/SystemCapabilityList';
import { TenantCapabilityAssignment } from './pages/capabilities/TenantCapabilityAssignment';
import { TenantFeatureEnablement } from './pages/capabilities/TenantFeatureEnablement';
import { UserCapabilityEnrollment } from './pages/capabilities/UserCapabilityEnrollment';

// Create QueryClient instance
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <PrincipalProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/login" element={<Login />} />
            <Route path="/no-access" element={<NoAccess />} />
            <Route
              path="/"
              element={
                <ProtectedRoute>
                  <Layout>
                    <Dashboard />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/tenants"
              element={
                <ProtectedRoute>
                  <Layout>
                    <TenantList />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/tenants/:id"
              element={
                <ProtectedRoute>
                  <Layout>
                    <TenantDetail />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/users"
              element={
                <ProtectedRoute>
                  <Layout>
                    <UserList />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/users/:id"
              element={
                <ProtectedRoute>
                  <Layout>
                    <UserDetail />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/roles"
              element={
                <ProtectedRoute>
                  <Layout>
                    <RoleList />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/permissions"
              element={
                <ProtectedRoute>
                  <Layout>
                    <PermissionList />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/settings"
              element={
                <ProtectedRoute>
                  <Layout>
                    <Settings />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/audit"
              element={
                <ProtectedRoute>
                  <Layout>
                    <AuditLogs />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/mfa"
              element={
                <ProtectedRoute>
                  <Layout>
                    <MFA />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/capabilities/system"
              element={
                <ProtectedRoute>
                  <Layout>
                    <SystemCapabilityList />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/capabilities/tenant-assignment"
              element={
                <ProtectedRoute>
                  <Layout>
                    <TenantCapabilityAssignment />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/capabilities/features"
              element={
                <ProtectedRoute>
                  <Layout>
                    <TenantFeatureEnablement />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route
              path="/capabilities/user-enrollment"
              element={
                <ProtectedRoute>
                  <Layout>
                    <UserCapabilityEnrollment />
                  </Layout>
                </ProtectedRoute>
              }
            />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </PrincipalProvider>
    </QueryClientProvider>
  );
}

export default App;
