"use client";

import Link from "next/link";
import { useAuth } from "@/components/auth/auth-provider";
import { Button } from "@/components/ui/button";
import { 
  BookmarkIcon, 
  UserIcon, 
  HomeIcon, 
  SettingsIcon, 
  LogOutIcon, 
  LogInIcon,
  TagIcon
} from "lucide-react";
import { usePathname } from "next/navigation";

export default function Header() {
  const { user, isAuthenticated, logout } = useAuth();
  const pathname = usePathname();
  
  return (
    <header className="sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-14 items-center justify-between">
        <div className="flex items-center gap-2">
          <Link href="/" className="flex items-center gap-2 font-semibold">
            <BookmarkIcon className="h-5 w-5" />
            <span>Memento</span>
          </Link>
          
          {isAuthenticated && (
            <nav className="ml-8 hidden md:flex gap-6">
              <Link 
                href="/articles" 
                className={`text-sm font-medium ${pathname === '/articles' ? 'text-primary' : 'text-muted-foreground hover:text-foreground'}`}
              >
                Articles
              </Link>
              <Link 
                href="/categories" 
                className={`text-sm font-medium ${pathname === '/categories' ? 'text-primary' : 'text-muted-foreground hover:text-foreground'}`}
              >
                Categories
              </Link>
              <Link 
                href="/tags" 
                className={`text-sm font-medium ${pathname === '/tags' ? 'text-primary' : 'text-muted-foreground hover:text-foreground'}`}
              >
                Tags
              </Link>
            </nav>
          )}
        </div>
        
        <div className="flex items-center gap-2">
          {isAuthenticated ? (
            <>
              <span className="hidden md:inline-block text-sm text-muted-foreground mr-2">
                {user?.name}
              </span>
              <Button 
                variant="ghost" 
                size="icon" 
                asChild
              >
                <Link href="/profile">
                  <UserIcon className="h-4 w-4" />
                </Link>
              </Button>
              <Button 
                variant="ghost" 
                size="icon"
                onClick={logout}
              >
                <LogOutIcon className="h-4 w-4" />
              </Button>
            </>
          ) : (
            <>
              <Button variant="ghost" asChild>
                <Link href="/login">
                  <LogInIcon className="mr-2 h-4 w-4" />
                  Login
                </Link>
              </Button>
              <Button variant="default" asChild>
                <Link href="/register">
                  Register
                </Link>
              </Button>
            </>
          )}
        </div>
      </div>
    </header>
  );
}