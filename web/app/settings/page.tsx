"use client";

import { useState, useEffect } from "react";
import { useTheme } from "next-themes";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { toast } from "@/hooks/use-toast";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  RadioGroup,
  RadioGroupItem,
} from "@/components/ui/radio-group";
import {
  GlobeIcon,
  Moon,
  PaletteIcon,
  Settings2Icon,
  SunIcon,
} from "lucide-react";
import { configAPI } from "@/lib/api-client";
import { Input } from "@/components/ui/input";
import { useAuth } from "@/components/auth/auth-provider";
import { useRouter } from "next/navigation";

export default function SettingsPage() {
  const { theme, setTheme } = useTheme();
  const { user } = useAuth();
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [appConfig, setAppConfig] = useState({
    bypass_refer: false,
    custom_scrape_retry_times: 3,
    custom_scrape_timeout_seconds: 30,
    custom_user_agent: "",
    custom_user_proxy: "",
    extract_links: true,
    llm_auto_gen_abstract: false,
    llm_auto_gen_tags: false,
    locale: "en",
    scrape_img_offline: true,
  });
  const [isLoading, setIsLoading] = useState(true);

  // If not logged in, redirect to login
  if (!user) {
    router.push("/login");
    return null;
  }

  // Fetch app configuration
  useEffect(() => {
    const fetchConfig = async () => {
      try {
        setIsLoading(true);
        const response = await configAPI.getConfig();
        if (response.data) {
          setAppConfig({
            bypass_refer: response.data.bypass_refer ?? false,
            custom_scrape_retry_times: response.data.custom_scrape_retry_times ?? 3,
            custom_scrape_timeout_seconds: response.data.custom_scrape_timeout_seconds ?? 30,
            custom_user_agent: response.data.custom_user_agent ?? "",
            custom_user_proxy: response.data.custom_user_proxy ?? "",
            extract_links: response.data.extract_links ?? true,
            llm_auto_gen_abstract: response.data.llm_auto_gen_abstract ?? false,
            llm_auto_gen_tags: response.data.llm_auto_gen_tags ?? false,
            locale: response.data.locale ?? "en",
            scrape_img_offline: response.data.scrape_img_offline ?? true,
          });
        }
      } catch (error) {
        console.error("Failed to fetch app configuration:", error);
        toast({
          title: "Error",
          description: "Failed to load your settings. Please try again.",
          variant: "destructive",
        });
      } finally {
        setIsLoading(false);
      }
    };

    fetchConfig();
  }, []);

  const handleToggleChange = (field: string) => {
    setAppConfig((prev) => ({
      ...prev,
      [field]: !prev[field as keyof typeof prev],
    }));
  };

  const handleInputChange = (field: string, value: string | number) => {
    setAppConfig((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleThemeChange = (value: string) => {
    setTheme(value);
  };

  const saveSettings = async () => {
    setIsSubmitting(true);
    try {
      await configAPI.updateConfig(appConfig);
      
      toast({
        title: "Settings saved",
        description: "Your settings have been updated successfully.",
      });
    } catch (error) {
      console.error("Failed to update settings:", error);
      toast({
        title: "Error",
        description: "Failed to save your settings. Please try again.",
        variant: "destructive",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isLoading) {
    return (
      <div className="container mx-auto max-w-3xl px-4 py-8">
        <p>Loading settings...</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto max-w-3xl px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="mt-2 text-muted-foreground">
          Customize your Memento experience
        </p>
      </div>

      <Tabs defaultValue="appearance">
        <TabsList className="mb-6 grid w-full grid-cols-3">
          <TabsTrigger value="appearance">Appearance</TabsTrigger>
          <TabsTrigger value="scraping">Web Scraping</TabsTrigger>
          <TabsTrigger value="ai">AI Features</TabsTrigger>
        </TabsList>
        
        <TabsContent value="appearance">
          <Card>
            <CardHeader className="flex flex-row items-center gap-4">
              <PaletteIcon className="h-8 w-8 text-primary" />
              <div>
                <CardTitle>Appearance</CardTitle>
                <CardDescription>
                  Customize how Memento looks and feels
                </CardDescription>
              </div>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <h3 className="text-lg font-medium">Theme</h3>
                <RadioGroup
                  defaultValue={theme}
                  onValueChange={handleThemeChange}
                  className="grid grid-cols-3 gap-4"
                >
                  <Label
                    htmlFor="theme-light"
                    className="flex flex-col items-center gap-2 rounded-md border border-border p-4 cursor-pointer hover:bg-accent"
                  >
                    <RadioGroupItem
                      value="light"
                      id="theme-light"
                      className="sr-only"
                    />
                    <SunIcon className="h-6 w-6" />
                    <span>Light</span>
                  </Label>
                  <Label
                    htmlFor="theme-dark"
                    className="flex flex-col items-center gap-2 rounded-md border border-border p-4 cursor-pointer hover:bg-accent"
                  >
                    <RadioGroupItem
                      value="dark"
                      id="theme-dark"
                      className="sr-only"
                    />
                    <Moon className="h-6 w-6" />
                    <span>Dark</span>
                  </Label>
                  <Label
                    htmlFor="theme-system"
                    className="flex flex-col items-center gap-2 rounded-md border border-border p-4 cursor-pointer hover:bg-accent"
                  >
                    <RadioGroupItem
                      value="system"
                      id="theme-system"
                      className="sr-only"
                    />
                    <Settings2Icon className="h-6 w-6" />
                    <span>System</span>
                  </Label>
                </RadioGroup>
              </div>

              <div className="space-y-4">
                <h3 className="text-lg font-medium">Language</h3>
                <div className="flex items-center gap-4">
                  <GlobeIcon className="h-5 w-5 text-muted-foreground" />
                  <div className="grid w-full max-w-sm items-center gap-1.5">
                    <Label htmlFor="locale">Locale</Label>
                    <Input 
                      id="locale" 
                      value={appConfig.locale}
                      onChange={(e) => handleInputChange('locale', e.target.value)}
                      maxLength={10}
                    />
                  </div>
                </div>
              </div>

              <Button onClick={saveSettings} disabled={isSubmitting}>
                {isSubmitting ? "Saving..." : "Save changes"}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
        
        <TabsContent value="scraping">
          <Card>
            <CardHeader className="flex flex-row items-center gap-4">
              <Settings2Icon className="h-8 w-8 text-primary" />
              <div>
                <CardTitle>Web Scraping</CardTitle>
                <CardDescription>
                  Configure scraping behavior for article collection
                </CardDescription>
              </div>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="scrape-img-offline">Offline Images</Label>
                    <p className="text-sm text-muted-foreground">
                      Save images locally when scraping articles
                    </p>
                  </div>
                  <Switch
                    id="scrape-img-offline"
                    checked={appConfig.scrape_img_offline}
                    onCheckedChange={() => handleToggleChange("scrape_img_offline")}
                  />
                </div>
                
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="bypass-refer">Bypass Referrer</Label>
                    <p className="text-sm text-muted-foreground">
                      Do not send referrer header when fetching content
                    </p>
                  </div>
                  <Switch
                    id="bypass-refer"
                    checked={appConfig.bypass_refer}
                    onCheckedChange={() => handleToggleChange("bypass_refer")}
                  />
                </div>
                
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="extract-links">Extract Links</Label>
                    <p className="text-sm text-muted-foreground">
                      Extract and store links from articles
                    </p>
                  </div>
                  <Switch
                    id="extract-links"
                    checked={appConfig.extract_links}
                    onCheckedChange={() => handleToggleChange("extract_links")}
                  />
                </div>

                <div className="space-y-4 pt-4">
                  <h3 className="text-lg font-medium">Advanced Settings</h3>
                  
                  <div className="grid w-full max-w-sm items-center gap-1.5">
                    <Label htmlFor="custom-user-agent">Custom User Agent</Label>
                    <Input 
                      id="custom-user-agent" 
                      value={appConfig.custom_user_agent}
                      onChange={(e) => handleInputChange('custom_user_agent', e.target.value)}
                    />
                  </div>
                  
                  <div className="grid w-full max-w-sm items-center gap-1.5">
                    <Label htmlFor="custom-user-proxy">Custom Proxy URL</Label>
                    <Input 
                      id="custom-user-proxy" 
                      value={appConfig.custom_user_proxy}
                      onChange={(e) => handleInputChange('custom_user_proxy', e.target.value)}
                    />
                  </div>
                  
                  <div className="grid w-full max-w-sm items-center gap-1.5">
                    <Label htmlFor="custom-retry-times">Retry Times</Label>
                    <Input 
                      id="custom-retry-times" 
                      type="number"
                      value={appConfig.custom_scrape_retry_times}
                      onChange={(e) => handleInputChange('custom_scrape_retry_times', parseInt(e.target.value))}
                    />
                  </div>
                  
                  <div className="grid w-full max-w-sm items-center gap-1.5">
                    <Label htmlFor="custom-timeout">Timeout (seconds)</Label>
                    <Input 
                      id="custom-timeout" 
                      type="number"
                      value={appConfig.custom_scrape_timeout_seconds}
                      onChange={(e) => handleInputChange('custom_scrape_timeout_seconds', parseInt(e.target.value))}
                    />
                  </div>
                </div>
              </div>

              <Button onClick={saveSettings} disabled={isSubmitting}>
                {isSubmitting ? "Saving..." : "Save changes"}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
        
        <TabsContent value="ai">
          <Card>
            <CardHeader className="flex flex-row items-center gap-4">
              <Settings2Icon className="h-8 w-8 text-primary" />
              <div>
                <CardTitle>AI Features</CardTitle>
                <CardDescription>
                  Configure AI assistance for article processing
                </CardDescription>
              </div>
            </CardHeader>
            <CardContent className="space-y-6">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="llm-auto-gen-abstract">Auto-generate Abstracts</Label>
                    <p className="text-sm text-muted-foreground">
                      Use AI to automatically generate article summaries
                    </p>
                  </div>
                  <Switch
                    id="llm-auto-gen-abstract"
                    checked={appConfig.llm_auto_gen_abstract}
                    onCheckedChange={() => handleToggleChange("llm_auto_gen_abstract")}
                  />
                </div>
                
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="llm-auto-gen-tags">Auto-generate Tags</Label>
                    <p className="text-sm text-muted-foreground">
                      Use AI to automatically generate relevant tags
                    </p>
                  </div>
                  <Switch
                    id="llm-auto-gen-tags"
                    checked={appConfig.llm_auto_gen_tags}
                    onCheckedChange={() => handleToggleChange("llm_auto_gen_tags")}
                  />
                </div>
              </div>

              <Button onClick={saveSettings} disabled={isSubmitting}>
                {isSubmitting ? "Saving..." : "Save changes"}
              </Button>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}