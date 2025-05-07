"use client";

import { useState, useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { useToast } from "@/components/ui/use-toast";
import { useAuth } from "@/components/auth/auth-provider";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { CalendarIcon, UserIcon } from "lucide-react";
import { format, parseISO } from "date-fns";
import { userAPI } from "@/lib/api-client";

const profileSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  email: z.string().email("Please enter a valid email address"),
});

type ProfileFormValues = z.infer<typeof profileSchema>;

export default function ProfilePage() {
  const { user, isAuthenticated, isLoading } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [activeTab, setActiveTab] = useState("profile");
  const { toast } = useToast();
  
  // Define form even if user is null initially - we'll update it when user loads
  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      name: user?.name || "",
      email: user?.email || "",
    },
  });

  // Update form values when user data changes
  useEffect(() => {
    if (user) {
      form.reset({
        name: user.name,
        email: user.email,
      });
    }
  }, [user, form]);
  
  // Set active tab from URL params
  useEffect(() => {
    const tab = searchParams.get("tab");
    if (tab === "account") {
      setActiveTab("account");
    }
  }, [searchParams]);

  // No longer redirecting from profile page - auth provider handles this
  // This prevents redirect loops and allows the profile page to render
  
  const onSubmit = async (data: ProfileFormValues) => {
    setIsSubmitting(true);
    try {
      // Track whether any updates were made
      let updatedFields = [];
      
      // Update user information
      if (user && data.name !== user.name) {
        await userAPI.updateName(data.name);
        updatedFields.push("name");
      }
      
      if (user && data.email !== user.email) {
        await userAPI.updateEmail(data.email);
        updatedFields.push("email");
      }
      
      // Update locally stored user information
      const storedUser = localStorage.getItem("mementoUser");
      if (storedUser) {
        const userData = JSON.parse(storedUser);
        userData.name = data.name;
        userData.email = data.email;
        localStorage.setItem("mementoUser", JSON.stringify(userData));
      }
      
      // Only show toast if something was actually updated
      if (updatedFields.length > 0) {
        toast({
          title: "Profile updated",
          description: `Your profile ${updatedFields.join(" and ")} has been successfully updated.`,
        });
      }
    } catch (error) {
      console.error("Error updating profile:", error);
      toast({
        title: "Error",
        description: "Failed to update profile. Please try again.",
        variant: "destructive",
      });
    } finally {
      setIsSubmitting(false);
    }
  };
  
  // Loading and authentication states with more detailed messages
  if (isLoading) {
    return (
      <div className="container mx-auto max-w-3xl px-4 py-8 flex justify-center items-center min-h-[50vh]">
        <p className="animate-pulse">Loading user data...</p>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="container mx-auto max-w-3xl px-4 py-8 flex justify-center items-center min-h-[50vh]">
        <p>Please log in to view your profile. Redirecting to login...</p>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="container mx-auto max-w-3xl px-4 py-8 flex justify-center items-center min-h-[50vh]">
        <p>Loading user profile data...</p>
      </div>
    );
  }

  // Main render content
  return (
    <div className="container mx-auto max-w-3xl px-4 py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold">Profile</h1>
        <p className="mt-2 text-muted-foreground">
          Manage your account information
        </p>
      </div>

      <Tabs defaultValue={activeTab} value={activeTab} onValueChange={setActiveTab}>
        <TabsList className="mb-6 grid w-full grid-cols-2 md:w-[400px]">
          <TabsTrigger value="profile">Profile</TabsTrigger>
          <TabsTrigger value="account">Account</TabsTrigger>
        </TabsList>
        
        <TabsContent value="profile">
          <Card>
            <CardHeader>
              <CardTitle>Profile Information</CardTitle>
              <CardDescription>
                Update your profile information and email address.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Form {...form}>
                <form
                  onSubmit={form.handleSubmit(onSubmit)}
                  className="space-y-6"
                >
                  <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Name</FormLabel>
                        <FormControl>
                          <Input {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  
                  <FormField
                    control={form.control}
                    name="email"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Email</FormLabel>
                        <FormControl>
                          <Input {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  
                  <Button type="submit" disabled={isSubmitting}>
                    {isSubmitting ? "Saving..." : "Save Changes"}
                  </Button>
                </form>
              </Form>
            </CardContent>
          </Card>
        </TabsContent>
        
        <TabsContent value="account">
          <Card>
            <CardHeader>
              <CardTitle>Account Information</CardTitle>
              <CardDescription>
                View your account details and membership information.
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex flex-col space-y-4">
                <div className="flex items-center gap-3 rounded-md border p-3">
                  <UserIcon className="h-5 w-5 text-muted-foreground" />
                  <div>
                    <div className="text-sm font-medium">Account Type</div>
                    <div className="text-sm text-muted-foreground">
                      {user?.role === "root" ? "Administrator" : "Regular User"}
                    </div>
                  </div>
                </div>
                
                <div className="flex items-center gap-3 rounded-md border p-3">
                  <CalendarIcon className="h-5 w-5 text-muted-foreground" />
                  <div>
                    <div className="text-sm font-medium">Registration Date</div>
                    <div className="text-sm text-muted-foreground">
                      {(() => {
                        if (user?.created_at) {
                          try {
                            return format(new Date(user.created_at), "MMMM d, yyyy");
                          } catch (error) {
                            console.error("Error formatting date:", error, user.created_at);
                            return "Invalid date format";
                          }
                        } else {
                          return "Not available";
                        }
                      })()}
                    </div>
                  </div>
                </div>
              </div>
              
              {/* Password Reset Form */}
              <div className="mt-6 pt-6 border-t">
                <h3 className="mb-4 text-lg font-medium">Change Password</h3>
                <PasswordResetForm />
              </div>
              
              <div className="mt-6 pt-6 border-t">
                <h3 className="mb-4 text-lg font-medium">Danger Zone</h3>
                <Button variant="destructive">Delete Account</Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

// Password Reset Form Component
const passwordResetSchema = z.object({
  oldPassword: z.string().min(6, "Old password must be at least 6 characters"),
  newPassword: z.string().min(6, "New password must be at least 6 characters"),
  confirmPassword: z.string().min(6, "Password must be at least 6 characters")
}).refine(data => data.newPassword === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"]
});

type PasswordResetFormValues = z.infer<typeof passwordResetSchema>;

function PasswordResetForm() {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { toast } = useToast();
  
  // Always define form at the top of the component to avoid React Hooks errors
  const form = useForm<PasswordResetFormValues>({
    resolver: zodResolver(passwordResetSchema),
    defaultValues: {
      oldPassword: "",
      newPassword: "",
      confirmPassword: ""
    }
  });
  
  const onSubmit = async (data: PasswordResetFormValues) => {
    setIsSubmitting(true);
    try {
      await userAPI.updatePassword(data.oldPassword, data.newPassword);
      
      toast({
        title: "Password updated",
        description: "Your password has been successfully updated.",
      });
      
      // Reset form after successful submission
      form.reset();
    } catch (error) {
      console.error("Error updating password:", error);
      toast({
        title: "Error",
        description: "Failed to update password. Please make sure your old password is correct.",
        variant: "destructive",
      });
    } finally {
      setIsSubmitting(false);
    }
  };
  
  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="oldPassword"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Current Password</FormLabel>
              <FormControl>
                <Input type="password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        
        <FormField
          control={form.control}
          name="newPassword"
          render={({ field }) => (
            <FormItem>
              <FormLabel>New Password</FormLabel>
              <FormControl>
                <Input type="password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        
        <FormField
          control={form.control}
          name="confirmPassword"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Confirm New Password</FormLabel>
              <FormControl>
                <Input type="password" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Updating..." : "Update Password"}
        </Button>
      </form>
    </Form>
  );
}