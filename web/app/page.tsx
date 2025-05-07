import Link from "next/link";
import { Button } from "@/components/ui/button";
import { ArrowRightIcon, BookmarkIcon, HashIcon, LayersIcon, SearchIcon } from "lucide-react";

export default function Home() {
  return (
    <div className="container mx-auto px-4 py-8 md:py-16">
      {/* Hero Section */}
      <section className="mx-auto max-w-5xl text-center">
        <div className="space-y-4">
          <h1 className="text-4xl font-bold tracking-tight sm:text-5xl md:text-6xl">
            Save what matters for{" "}
            <span className="bg-gradient-to-r from-blue-600 to-indigo-600 bg-clip-text text-transparent">
              later
            </span>
          </h1>
          <p className="mx-auto max-w-3xl text-lg text-muted-foreground sm:text-xl">
            Memento helps you save and organize articles, videos, and more from
            around the web so you can read them later.
          </p>
          <div className="mt-8 flex flex-wrap justify-center gap-4">
            <Button size="lg" asChild>
              <Link href="/articles">
                Get Started
                <ArrowRightIcon className="ml-2 h-4 w-4" />
              </Link>
            </Button>
            <Button size="lg" variant="outline" asChild>
              <Link href="/login">Sign In</Link>
            </Button>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-16">
        <h2 className="mb-12 text-center text-3xl font-bold">How it works</h2>
        <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
          <div className="flex flex-col items-center text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary/10 text-primary mb-4">
              <BookmarkIcon className="h-8 w-8" />
            </div>
            <h3 className="text-xl font-semibold">Save</h3>
            <p className="mt-2 text-muted-foreground">
              Save articles, videos, and pages from any device with a simple URL.
            </p>
          </div>
          <div className="flex flex-col items-center text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary/10 text-primary mb-4">
              <HashIcon className="h-8 w-8" />
            </div>
            <h3 className="text-xl font-semibold">Organize</h3>
            <p className="mt-2 text-muted-foreground">
              Tag and categorize your content for easy filtering and discovery.
            </p>
          </div>
          <div className="flex flex-col items-center text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-primary/10 text-primary mb-4">
              <SearchIcon className="h-8 w-8" />
            </div>
            <h3 className="text-xl font-semibold">Find</h3>
            <p className="mt-2 text-muted-foreground">
              Quickly search and filter your saved content whenever you need it.
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}