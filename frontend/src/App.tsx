import React from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider, useAuth } from './context/AuthContext';
import { LanguageProvider } from './context/LanguageContext';
import { Navbar } from './components/layout/Navbar';
import { Sidebar } from './components/layout/Sidebar';
import { MobileNav } from './components/layout/MobileNav';
import { AuthModal } from './components/auth/AuthModal';

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
import { ShopPage } from './pages/ShopPage';
import { AnalyticsPage } from './pages/AnalyticsPage';

const queryClient = new QueryClient();

// Protected Route Wrapper
const ProtectedRoute: React.FC<{ children: React.ReactNode; roles?: string[] }> = ({ children, roles }) => {
  const { user, loading, openAuthModal } = useAuth();
  const location = useLocation();

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center text-sm font-semibold text-gray-500">
        Загрузка...
      </div>
    );
  }

  if (!user) {
    openAuthModal(location.pathname);
    return <Navigate to="/" replace />;
  }

  if (roles && !roles.includes(user.role)) {
    return <Navigate to="/dashboard" replace />;
  }

  return <>{children}</>;
};

const AppLayout: React.FC = () => {
  const location = useLocation();
  const isAuthPage = location.pathname === '/login' || location.pathname === '/register';
  const isAboutPage = location.pathname === '/about';

  // 1. Auth pages: standalone centered view on #F4F6F9 without Sidebar or Navbar
  if (isAuthPage) {
    return (
      <div className="min-h-screen bg-[#F5F8FD] text-[#1C2D4B] font-sans antialiased flex flex-col justify-center">
        <main className="w-full">
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
          </Routes>
        </main>
      </div>
    );
  }

  // 2. Landing / About page: standalone full-width presentation page
  if (isAboutPage) {
    return (
      <div className="min-h-screen bg-white text-[#1C2D4B] font-sans antialiased">
        <Routes>
          <Route path="/about" element={<LandingPage />} />
        </Routes>
      </div>
    );
  }

  // 3. Platform screens: deep navy sidebar + static navbar + bottom-padded main for mobile dock
  return (
    <div className="min-h-screen flex bg-[#F5F8FD] text-[#1C2D4B] font-sans antialiased">
      <Sidebar />

      <div className="flex-1 flex flex-col min-w-0">
        <Navbar />

        <main className="flex-1 p-3 sm:p-5 lg:p-7 max-w-[1400px] w-full mx-auto pb-24 md:pb-6">
          <Routes>
            {/* The root path / directly renders the dashboard */}
            <Route path="/" element={<DashboardPage />} />
            <Route path="/dashboard" element={<DashboardPage />} />

            {/* Core learning & task */}
            <Route path="/task" element={<DailyTaskPage />} />
            <Route path="/treasury" element={<TreasuryPage />} />
            <Route path="/leaderboard" element={<LeaderboardPage />} />

            {/* Authenticated routes */}
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
              path="/analytics"
              element={<AnalyticsPage />}
            />
            <Route
              path="/shop"
              element={
                <ProtectedRoute>
                  <ShopPage />
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

      <AuthModal />
    </div>
  );
};

export function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <LanguageProvider>
        <AuthProvider>
          <BrowserRouter>
            <AppLayout />
          </BrowserRouter>
        </AuthProvider>
      </LanguageProvider>
    </QueryClientProvider>
  );
}

export default App;
