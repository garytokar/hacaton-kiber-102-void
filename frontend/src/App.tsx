import React from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider, useAuth } from './context/AuthContext';
import { Navbar } from './components/layout/Navbar';
import { Sidebar } from './components/layout/Sidebar';
import { MobileNav } from './components/layout/MobileNav';

// Pages
import { LandingPage } from './pages/LandingPage';
import { LoginPage } from './pages/LoginPage';
import { RegisterPage } from './pages/RegisterPage';
import { DashboardPage } from './pages/DashboardPage';
import { DailyTaskPage } from './pages/DailyTaskPage';
import { TreasuryPage } from './pages/TreasuryPage';
import { DictionaryPage } from './pages/DictionaryPage';
import { LeaderboardPage } from './pages/LeaderboardPage';
import { AchievementsPage } from './pages/AchievementsPage';
import { ExpeditionsPage } from './pages/ExpeditionsPage';
import { ProfilePage } from './pages/ProfilePage';
import { AdminPage } from './pages/AdminPage';

const queryClient = new QueryClient();

// Protected Route Wrapper
const ProtectedRoute: React.FC<{ children: React.ReactNode; roles?: string[] }> = ({ children, roles }) => {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center text-sm font-semibold text-gray-500">
        Загрузка...
      </div>
    );
  }

  if (!user) {
    return <Navigate to="/login" replace />;
  }

  if (roles && !roles.includes(user.role)) {
    return <Navigate to="/dashboard" replace />;
  }

  return <>{children}</>;
};

const AppLayout: React.FC = () => {
  const { user } = useAuth();
  const location = useLocation();
  const isAuthPage = location.pathname === '/login' || location.pathname === '/register';

  // In the Figma design, the deep navy sidebar is fixed on the left for all platform screens
  const showSidebar = !isAuthPage;

  return (
    <div className="min-h-screen flex bg-[#F4F6F9] text-[#0C2340] font-sans antialiased">
      {showSidebar && <Sidebar />}

      <div className="flex-1 flex flex-col min-w-0">
        <Navbar />

        <main className="flex-1 p-3 sm:p-5 lg:p-6 max-w-[1400px] w-full mx-auto">
          <Routes>
            <Route path="/" element={<LandingPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />

            {/* Core learning & task */}
            <Route path="/task" element={<DailyTaskPage />} />
            <Route path="/treasury" element={<TreasuryPage />} />
            <Route path="/leaderboard" element={<LeaderboardPage />} />

            {/* Authenticated routes */}
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <DashboardPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/dictionary"
              element={
                <ProtectedRoute>
                  <DictionaryPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/expeditions"
              element={
                <ProtectedRoute>
                  <ExpeditionsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/achievements"
              element={
                <ProtectedRoute>
                  <AchievementsPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/profile"
              element={
                <ProtectedRoute>
                  <ProfilePage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/admin"
              element={
                <ProtectedRoute roles={['admin', 'expert', 'teacher']}>
                  <AdminPage />
                </ProtectedRoute>
              }
            />

            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </main>

        <MobileNav />
      </div>
    </div>
  );
};

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <BrowserRouter>
          <AppLayout />
        </BrowserRouter>
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;
