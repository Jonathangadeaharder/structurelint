import * as vscode from 'vscode';
import * as child_process from 'child_process';
import * as path from 'path';

let diagnosticCollection: vscode.DiagnosticCollection;
let outputChannel: vscode.OutputChannel;
let statusBarItem: vscode.StatusBarItem;

export function activate(context: vscode.ExtensionContext) {
	console.log('structurelint extension is now active');

	// Create diagnostic collection
	diagnosticCollection = vscode.languages.createDiagnosticCollection('structurelint');
	context.subscriptions.push(diagnosticCollection);

	// Create output channel
	outputChannel = vscode.window.createOutputChannel('structurelint');
	context.subscriptions.push(outputChannel);

	// Create status bar item
	statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left);
	statusBarItem.text = '$(check) structurelint';
	statusBarItem.command = 'structurelint.lint';
	statusBarItem.show();
	context.subscriptions.push(statusBarItem);

	// Register commands
	context.subscriptions.push(
		vscode.commands.registerCommand('structurelint.lint', () => lintWorkspace())
	);

	context.subscriptions.push(
		vscode.commands.registerCommand('structurelint.lintFile', () => lintCurrentFile())
	);

	context.subscriptions.push(
		vscode.commands.registerCommand('structurelint.fix', () => fixViolations())
	);

	context.subscriptions.push(
		vscode.commands.registerCommand('structurelint.exportGraph', () => exportGraph())
	);

	// Register event handlers
	const config = vscode.workspace.getConfiguration('structurelint');

	if (config.get('lintOnSave')) {
		context.subscriptions.push(
			vscode.workspace.onDidSaveTextDocument((document) => {
				if (isLintableDocument(document)) {
					lintWorkspace();
				}
			})
		);
	}

	if (config.get('lintOnOpen')) {
		context.subscriptions.push(
			vscode.workspace.onDidOpenTextDocument((document) => {
				if (isLintableDocument(document)) {
					lintWorkspace();
				}
			})
		);
	}

	// Lint on startup
	if (config.get('enable')) {
		lintWorkspace();
	}
}

export function deactivate() {
	if (diagnosticCollection) {
		diagnosticCollection.dispose();
	}
	if (outputChannel) {
		outputChannel.dispose();
	}
	if (statusBarItem) {
		statusBarItem.dispose();
	}
}

function isLintableDocument(document: vscode.TextDocument): boolean {
	// Skip output, git, and other special schemes
	return document.uri.scheme === 'file';
}

async function lintWorkspace() {
	const config = vscode.workspace.getConfiguration('structurelint');

	if (!config.get('enable')) {
		return;
	}

	const workspaceFolders = vscode.workspace.workspaceFolders;
	if (!workspaceFolders || workspaceFolders.length === 0) {
		vscode.window.showErrorMessage('No workspace folder open');
		return;
	}

	const workspaceRoot = workspaceFolders[0].uri.fsPath;

	statusBarItem.text = '$(sync~spin) structurelint: Running...';

	try {
		const violations = await runStructurelint(workspaceRoot);
		updateDiagnostics(violations, workspaceRoot);

		if (violations.length === 0) {
			statusBarItem.text = '$(check) structurelint: No issues';
			vscode.window.showInformationMessage('structurelint: All checks passed');
		} else {
			statusBarItem.text = `$(warning) structurelint: ${violations.length} issue(s)`;
		}
	} catch (error: any) {
		statusBarItem.text = '$(error) structurelint: Error';
		vscode.window.showErrorMessage(`structurelint error: ${error.message}`);
		outputChannel.appendLine(`Error: ${error.message}`);
	}
}

async function lintCurrentFile() {
	const editor = vscode.window.activeTextEditor;
	if (!editor) {
		vscode.window.showErrorMessage('No active editor');
		return;
	}

	const document = editor.document;
	if (!isLintableDocument(document)) {
		return;
	}

	await lintWorkspace(); // For now, just run full workspace lint
}

async function fixViolations() {
	const config = vscode.workspace.getConfiguration('structurelint');

	if (!config.get('enable')) {
		return;
	}

	const workspaceFolders = vscode.workspace.workspaceFolders;
	if (!workspaceFolders || workspaceFolders.length === 0) {
		vscode.window.showErrorMessage('No workspace folder open');
		return;
	}

	const workspaceRoot = workspaceFolders[0].uri.fsPath;

	const answer = await vscode.window.showWarningMessage(
		'Run structurelint --fix? This will modify your files.',
		'Yes',
		'Dry Run',
		'Cancel'
	);

	if (answer === 'Cancel' || !answer) {
		return;
	}

	const dryRun = answer === 'Dry Run';

	try {
		const output = await runFix(workspaceRoot, dryRun);
		outputChannel.clear();
		outputChannel.appendLine(output);
		outputChannel.show();

		if (dryRun) {
			vscode.window.showInformationMessage('structurelint: Dry run completed (no changes made)');
		} else {
			vscode.window.showInformationMessage('structurelint: Fixes applied successfully');
			// Re-lint after fixing
			await lintWorkspace();
		}
	} catch (error: any) {
		vscode.window.showErrorMessage(`structurelint --fix error: ${error.message}`);
		outputChannel.appendLine(`Error: ${error.message}`);
	}
}

async function exportGraph() {
	const config = vscode.workspace.getConfiguration('structurelint');

	if (!config.get('enable')) {
		return;
	}

	const workspaceFolders = vscode.workspace.workspaceFolders;
	if (!workspaceFolders || workspaceFolders.length === 0) {
		vscode.window.showErrorMessage('No workspace folder open');
		return;
	}

	const workspaceRoot = workspaceFolders[0].uri.fsPath;

	const format = await vscode.window.showQuickPick(['mermaid', 'dot', 'json'], {
		placeHolder: 'Select export format'
	});

	if (!format) {
		return;
	}

	try {
		const graph = await runExportGraph(workspaceRoot, format);

		// Create new document with graph
		const doc = await vscode.workspace.openTextDocument({
			content: graph,
			language: format === 'json' ? 'json' : format === 'mermaid' ? 'markdown' : 'dot'
		});

		await vscode.window.showTextDocument(doc);
		vscode.window.showInformationMessage(`structurelint: Dependency graph exported (${format})`);
	} catch (error: any) {
		vscode.window.showErrorMessage(`structurelint --export-graph error: ${error.message}`);
		outputChannel.appendLine(`Error: ${error.message}`);
	}
}

interface Violation {
	rule: string;
	path: string;
	message: string;
}

function runStructurelint(workspaceRoot: string): Promise<Violation[]> {
	return new Promise((resolve, reject) => {
		const config = vscode.workspace.getConfiguration('structurelint');
		const executablePath = config.get<string>('executablePath', 'structurelint');
		const additionalArgs = config.get<string[]>('additionalArgs', []);
		const productionMode = config.get<boolean>('productionMode', false);

		const args = [workspaceRoot, ...additionalArgs];
		if (productionMode) {
			args.push('--production');
		}

		outputChannel.appendLine(`Running: ${executablePath} ${args.join(' ')}`);

		const process = child_process.spawn(executablePath, args, {
			cwd: workspaceRoot
		});

		let stdout = '';
		let stderr = '';

		process.stdout.on('data', (data) => {
			stdout += data.toString();
		});

		process.stderr.on('data', (data) => {
			stderr += data.toString();
		});

		process.on('close', (code) => {
			outputChannel.appendLine(`Exit code: ${code}`);

			if (code === 0) {
				// No violations
				resolve([]);
			} else if (code === 1) {
				// Violations found, parse them
				const violations = parseViolations(stdout + stderr);
				resolve(violations);
			} else {
				// Error occurred
				reject(new Error(stderr || stdout || 'Unknown error'));
			}
		});

		process.on('error', (error) => {
			reject(new Error(`Failed to spawn structurelint: ${error.message}. Is it installed and in PATH?`));
		});
	});
}

function runFix(workspaceRoot: string, dryRun: boolean): Promise<string> {
	return new Promise((resolve, reject) => {
		const config = vscode.workspace.getConfiguration('structurelint');
		const executablePath = config.get<string>('executablePath', 'structurelint');

		const args = [workspaceRoot, '--fix'];
		if (dryRun) {
			args.push('--dry-run');
		}

		const process = child_process.spawn(executablePath, args, {
			cwd: workspaceRoot
		});

		let output = '';

		process.stdout.on('data', (data) => {
			output += data.toString();
		});

		process.stderr.on('data', (data) => {
			output += data.toString();
		});

		process.on('close', (code) => {
			if (code === 0) {
				resolve(output);
			} else {
				reject(new Error(output || 'Fix command failed'));
			}
		});

		process.on('error', (error) => {
			reject(new Error(`Failed to spawn structurelint: ${error.message}`));
		});
	});
}

function runExportGraph(workspaceRoot: string, format: string): Promise<string> {
	return new Promise((resolve, reject) => {
		const config = vscode.workspace.getConfiguration('structurelint');
		const executablePath = config.get<string>('executablePath', 'structurelint');

		const args = [workspaceRoot, '--export-graph', format];

		const process = child_process.spawn(executablePath, args, {
			cwd: workspaceRoot
		});

		let output = '';
		let error = '';

		process.stdout.on('data', (data) => {
			output += data.toString();
		});

		process.stderr.on('data', (data) => {
			error += data.toString();
		});

		process.on('close', (code) => {
			if (code === 0) {
				resolve(output);
			} else {
				reject(new Error(error || 'Export graph command failed'));
			}
		});

		process.on('error', (err) => {
			reject(new Error(`Failed to spawn structurelint: ${err.message}`));
		});
	});
}

function parseViolations(output: string): Violation[] {
	const violations: Violation[] = [];
	const lines = output.split('\n');

	for (const line of lines) {
		// Match format: "path: message"
		const match = line.match(/^(.+?):\s+(.+)$/);
		if (match) {
			violations.push({
				rule: 'structurelint',  // We don't parse rule name from output yet
				path: match[1].trim(),
				message: match[2].trim()
			});
		}
	}

	return violations;
}

function updateDiagnostics(violations: Violation[], workspaceRoot: string) {
	// Clear existing diagnostics
	diagnosticCollection.clear();

	// Group violations by file
	const diagnosticsByFile = new Map<string, vscode.Diagnostic[]>();

	for (const violation of violations) {
		const filePath = path.isAbsolute(violation.path)
			? violation.path
			: path.join(workspaceRoot, violation.path);

		const uri = vscode.Uri.file(filePath);

		if (!diagnosticsByFile.has(filePath)) {
			diagnosticsByFile.set(filePath, []);
		}

		// Create diagnostic at line 0 (we don't parse line numbers yet)
		const range = new vscode.Range(0, 0, 0, 0);
		const diagnostic = new vscode.Diagnostic(
			range,
			violation.message,
			vscode.DiagnosticSeverity.Warning
		);

		diagnostic.source = 'structurelint';
		diagnostic.code = violation.rule;

		diagnosticsByFile.get(filePath)!.push(diagnostic);
	}

	// Set diagnostics for each file
	for (const [filePath, diagnostics] of diagnosticsByFile) {
		const uri = vscode.Uri.file(filePath);
		diagnosticCollection.set(uri, diagnostics);
	}
}
