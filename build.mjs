import * as esbuild from 'esbuild'
import {sassPlugin} from 'esbuild-sass-plugin'
import {readdirSync, statSync} from 'fs'
import {join} from 'path'
import { fileURLToPath } from 'url';
import { dirname } from 'path';

const cwd = dirname(fileURLToPath(import.meta.url));
const assetDirectory = join(cwd, "assets");

const filenames = readdirSync(assetDirectory).filter(file => {
    // Use fs.statSync to check if the item is a file
    return statSync(join(assetDirectory, file)).isFile();
}).map(file => join(assetDirectory, file));

const options = {
    entryPoints: filenames,
    bundle: true,
    outdir: 'static',
    plugins: [sassPlugin()],
    minify: true,
};

if (process.argv.includes('--watch')) {
    const context = await esbuild.context(options);
    await context.watch();
} else {
    await esbuild.build(options);
}

