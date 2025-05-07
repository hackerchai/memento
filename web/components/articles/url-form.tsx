"use client";

import { useState } from "react";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormMessage,
} from "@/components/ui/form";
import { toast } from "@/hooks/use-toast";
import { articleAPI } from "@/lib/api-client";

const urlSchema = z.object({
  url: z
    .string()
    .min(1, "URL is required")
    .url("Please enter a valid URL"),
});

type UrlFormValues = z.infer<typeof urlSchema>;

export function UrlForm() {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const router = useRouter();

  const form = useForm<UrlFormValues>({
    resolver: zodResolver(urlSchema),
    defaultValues: {
      url: "",
    },
  });

  const onSubmit = async (data: UrlFormValues) => {
    setIsSubmitting(true);
    try {
      // Use the article API client to create a new article
      await articleAPI.createArticle(data.url);
      
      toast({
        title: "Success!",
        description: "The article has been saved to your library.",
      });
      
      form.reset();
      router.refresh();
    } catch (error) {
      console.error("Error saving URL:", error);
      toast({
        title: "Something went wrong",
        description: error instanceof Error ? error.message : "Failed to save the article. Please try again.",
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
          name="url"
          render={({ field }) => (
            <FormItem>
              <FormControl>
                <div className="flex space-x-2">
                  <Input
                    placeholder="Paste a URL to save (e.g., https://example.com/article)"
                    {...field}
                    className="flex-1"
                  />
                  <Button type="submit" disabled={isSubmitting}>
                    {isSubmitting ? "Saving..." : "Save"}
                  </Button>
                </div>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </form>
    </Form>
  );
}