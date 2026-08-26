import { z } from 'zod';

export const ProfileSchema = z.object({
  address: z.string().default('http://localhost:8080'),
  namespace: z.string().default('default'),
  apiKey: z.string().optional(),
  tls: z.boolean().default(false),
  tlsCertPath: z.string().optional(),
  tlsKeyPath: z.string().optional(),
  tlsCaPath: z.string().optional(),
  codecEndpoint: z.string().optional(),
  output: z.enum(['text', 'json', 'yaml']).default('text'),
});

export type Profile = z.infer<typeof ProfileSchema>;

export const ConfigSchema = z.object({
  version: z.number().default(1),
  activeProfile: z.string().default('default'),
  profiles: z.record(ProfileSchema).default({}),
});

export type Config = z.infer<typeof ConfigSchema>;

export const GlobalFlagsSchema = z.object({
  address: z.string().optional(),
  namespace: z.string().optional(),
  apiKey: z.string().optional(),
  tls: z.boolean().optional(),
  tlsCertPath: z.string().optional(),
  tlsKeyPath: z.string().optional(),
  tlsCaPath: z.string().optional(),
  codecEndpoint: z.string().optional(),
  output: z.enum(['text', 'json', 'yaml']).optional(),
  color: z.enum(['always', 'never', 'auto']).optional(),
  profile: z.string().optional(),
  logLevel: z.enum(['debug', 'info', 'warn', 'error', 'never']).optional(),
});

export type GlobalFlags = z.infer<typeof GlobalFlagsSchema> & { _cmd?: import('commander').Command };
