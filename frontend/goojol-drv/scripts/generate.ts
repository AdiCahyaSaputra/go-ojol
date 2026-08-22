#!/usr/bin/env node

import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Helper functions to convert naming conventions
function toKebabCase(str: string): string {
  return str
    .replace(/([a-z])([A-Z])/g, '$1-$2')
    .replace(/[\s_]+/g, '-')
    .toLowerCase();
}

function toCamelCase(str: string): string {
  return str
    .replace(/[-_\s]+(.)?/g, (_, c) => (c ? c.toUpperCase() : ''))
    .replace(/^(.)/, (c) => c.toLowerCase());
}

function toPascalCase(str: string): string {
  return str
    .replace(/[-_\s]+(.)?/g, (_, c) => (c ? c.toUpperCase() : ''))
    .replace(/^(.)/, (c) => c.toUpperCase());
}

function toSingular(str: string): string {
  if (str.endsWith('s')) return str.slice(0, -1);
  if (str.endsWith('y')) return str.slice(0, -1) + 'y';
  return str;
}

function parseFeaturePath(name: string): { segments: string[]; baseName: string } {
  const segments = name
    .split('/')
    .map((segment) => segment.trim())
    .filter(Boolean)
    .map(toKebabCase);

  if (segments.length === 0) {
    throw new Error('Invalid feature name');
  }

  return {
    segments,
    baseName: segments[segments.length - 1],
  };
}

// Parse command line arguments
function parseArgs(): { name: string } | null {
  const args = process.argv.slice(2);
  let name = '';

  for (let i = 0; i < args.length; i++) {
    if (args[i] === '--name' && args[i + 1]) {
      name = args[i + 1];
      break;
    }
  }

  if (!name) {
    console.error('Error: --name parameter is required');
    console.log('Usage: node scripts/generate.ts --name <feature-name>');
    console.log('       node scripts/generate.ts --name <parent>/<sub-feature>');
    return null;
  }

  return { name };
}

// Replace all variations of "example" in content
function replaceContent(content: string, name: string): string {
  const kebabName = toKebabCase(name);
  const camelName = toCamelCase(name);
  const pascalName = toPascalCase(name);
  const pluralKebab = toSingular(kebabName);
  const pluralPascal = toSingular(pascalName);

  return content
    // Replace plural forms first to avoid partial replacements
    .replace(/examples/g, pluralKebab)
    .replace(/Examples/g, pluralPascal)
    // Replace singular forms
    .replace(/example-/g, `${kebabName}-`)
    .replace(/example\./g, `${kebabName}.`)
    .replace(/example"/g, `${kebabName}"`)
    .replace(/example]/g, `${kebabName}]`)
    .replace(/example\)/g, `${kebabName})`)
    .replace(/example\s/g, `${kebabName} `)
    .replace(/Example([A-Z])/g, (_, next) => `${pascalName}${next}`)
    .replace(/example([A-Z])/g, (_, next) => `${camelName}${next}`)
    .replace(/Example\s/g, `${pascalName} `)
    .replace(/Example$/g, pascalName)
    .replace(/example$/g, camelName);
}

// Replace example in filename
function replaceFilename(filename: string, name: string): string {
  const kebabName = toKebabCase(name);
  return filename.replace(/example/g, kebabName);
}

// Recursively copy and transform files
function processDirectory(
  sourceDir: string,
  targetDir: string,
  name: string,
  stats: { created: string[]; skipped: string[] }
): void {
  // Create target directory if it doesn't exist
  if (!fs.existsSync(targetDir)) {
    fs.mkdirSync(targetDir, { recursive: true });
  }

  const entries = fs.readdirSync(sourceDir, { withFileTypes: true });

  for (const entry of entries) {
    const sourcePath = path.join(sourceDir, entry.name);
    const targetFilename = replaceFilename(entry.name, name);
    const targetPath = path.join(targetDir, targetFilename);

    if (entry.isDirectory()) {
      // Recursively process subdirectories
      processDirectory(sourcePath, targetPath, name, stats);
    } else if (entry.isFile()) {
      // Check if file already exists
      if (fs.existsSync(targetPath)) {
        console.log(`⊘ Skipped (already exists): ${targetPath}`);
        stats.skipped.push(targetPath);
        continue;
      }

      // Read file content
      const content = fs.readFileSync(sourcePath, 'utf-8');
      
      // Replace content
      const newContent = replaceContent(content, name);
      
      // Write to target
      fs.writeFileSync(targetPath, newContent, 'utf-8');
      console.log(`✓ Created: ${targetPath}`);
      stats.created.push(targetPath);
    }
  }
}

// Main function
function main() {
  const args = parseArgs();
  if (!args) {
    process.exit(1);
  }

  const { name } = args;
  const { segments, baseName } = parseFeaturePath(name);
  const featurePath = segments.join('/');

  // Define paths
  const rootDir = path.resolve(__dirname, '..');
  const examplesDir = path.join(rootDir, 'scripts', 'examples');
  const featuresDir = path.join(rootDir, 'features');
  const targetDir = path.join(featuresDir, ...segments);

  // Check if examples directory exists
  if (!fs.existsSync(examplesDir)) {
    console.error(`Error: Examples directory not found at ${examplesDir}`);
    process.exit(1);
  }

  const featureExists = fs.existsSync(targetDir);
  
  console.log(`\n${featureExists ? 'Updating' : 'Generating'} feature: ${featurePath}`);
  console.log(`Source: ${examplesDir}`);
  console.log(`Target: ${targetDir}\n`);

  // Track created and skipped files
  const stats = { created: [] as string[], skipped: [] as string[] };

  // Process and copy files (baseName is the leaf segment used in filenames/content)
  processDirectory(examplesDir, targetDir, baseName, stats);

  console.log(`\n${featureExists ? '✅ Feature updated!' : '✅ Feature created successfully!'}\n`);
  console.log(`Location: src/features/${featurePath}/\n`);
  
  if (stats.created.length > 0) {
    console.log(`✓ Created ${stats.created.length} file(s):`);
    stats.created.forEach(file => {
      const relativePath = file.replace(targetDir + '/', '');
      console.log(`  - ${relativePath}`);
    });
  }
  
  if (stats.skipped.length > 0) {
    console.log(`\n⊘ Skipped ${stats.skipped.length} existing file(s):`);
    stats.skipped.forEach(file => {
      const relativePath = file.replace(targetDir + '/', '');
      console.log(`  - ${relativePath}`);
    });
  }
  
  if (stats.created.length === 0 && stats.skipped.length === 0) {
    console.log('No files to process.');
  }
}

main();

