import defaultMdxComponents from "fumadocs-ui/mdx";
import type { MDXComponents } from "mdx/types";
import { View } from "./docs-view-switch";

export function getMDXComponents(components?: MDXComponents) {
  return {
    ...defaultMdxComponents,
    View,
    ...components,
  } satisfies MDXComponents;
}

export const useMDXComponents = getMDXComponents;

declare global {
  type MDXProvidedComponents = ReturnType<typeof getMDXComponents>;
}
