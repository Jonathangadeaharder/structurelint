#!/usr/bin/env node
/**
 * JavaScript/TypeScript metrics calculator for structurelint.
 * Calculates cognitive complexity and Halstead metrics.
 *
 * Requires: @babel/parser for robust JS/TS parsing
 * Install: npm install -g @babel/parser
 */

const fs = require('fs');
const path = require('path');

// Try to load @babel/parser
let parser;
try {
    parser = require('@babel/parser');
} catch (e) {
    console.error(JSON.stringify({
        error: '@babel/parser not found. Please install with: npm install -g @babel/parser',
        functions: [],
        file_level: {}
    }));
    process.exit(1);
}

/**
 * Calculate cognitive complexity for a function or file
 */
class CognitiveComplexityCalculator {
    constructor() {
        this.complexity = 0;
        this.nestingLevel = 0;
        this.functionMetrics = [];
    }

    visitNode(node, nestingLevel = 0) {
        if (!node) return;

        // Handle function declarations
        if (this.isFunctionNode(node)) {
            this.visitFunction(node);
            return;
        }

        switch (node.type) {
            case 'IfStatement':
                this.complexity += 1 + nestingLevel;
                if (node.consequent) {
                    this.visitBlock(node.consequent, nestingLevel + 1);
                }
                // Handle else/else if - don't increase nesting for else if
                if (node.alternate) {
                    if (node.alternate.type === 'IfStatement') {
                        this.visitNode(node.alternate, nestingLevel);
                    } else {
                        this.visitBlock(node.alternate, nestingLevel + 1);
                    }
                }
                break;

            case 'ForStatement':
            case 'ForInStatement':
            case 'ForOfStatement':
            case 'WhileStatement':
            case 'DoWhileStatement':
                this.complexity += 1 + nestingLevel;
                if (node.body) {
                    this.visitBlock(node.body, nestingLevel + 1);
                }
                break;

            case 'SwitchStatement':
                this.complexity += 1 + nestingLevel;
                if (node.cases) {
                    node.cases.forEach(caseNode => {
                        if (caseNode.test) { // Not default case
                            this.complexity += 1;
                        }
                        caseNode.consequent.forEach(stmt => {
                            this.visitNode(stmt, nestingLevel + 1);
                        });
                    });
                }
                break;

            case 'TryStatement':
                this.complexity += 1 + nestingLevel;
                if (node.block) {
                    this.visitBlock(node.block, nestingLevel + 1);
                }
                if (node.handler) {
                    this.complexity += 1;
                    this.visitBlock(node.handler.body, nestingLevel + 1);
                }
                if (node.finalizer) {
                    this.visitBlock(node.finalizer, nestingLevel);
                }
                break;

            case 'CatchClause':
                // Handled in TryStatement
                break;

            case 'ConditionalExpression':
                this.complexity += 1 + nestingLevel;
                if (node.consequent) this.visitNode(node.consequent, nestingLevel);
                if (node.alternate) this.visitNode(node.alternate, nestingLevel);
                break;

            case 'LogicalExpression':
                if (node.operator === '&&' || node.operator === '||') {
                    this.complexity += 1;
                }
                if (node.left) this.visitNode(node.left, nestingLevel);
                if (node.right) this.visitNode(node.right, nestingLevel);
                break;

            case 'BreakStatement':
            case 'ContinueStatement':
                this.complexity += 1 + nestingLevel;
                break;

            case 'BlockStatement':
                node.body.forEach(stmt => this.visitNode(stmt, nestingLevel));
                break;

            case 'Program':
                node.body.forEach(stmt => this.visitNode(stmt, nestingLevel));
                break;

            default:
                // Visit child nodes
                this.visitChildren(node, nestingLevel);
                break;
        }
    }

    visitBlock(node, nestingLevel) {
        if (!node) return;

        if (node.type === 'BlockStatement') {
            node.body.forEach(stmt => this.visitNode(stmt, nestingLevel));
        } else {
            this.visitNode(node, nestingLevel);
        }
    }

    visitChildren(node, nestingLevel) {
        if (!node || typeof node !== 'object') return;

        Object.keys(node).forEach(key => {
            const child = node[key];
            if (Array.isArray(child)) {
                child.forEach(item => {
                    if (item && typeof item === 'object' && item.type) {
                        this.visitNode(item, nestingLevel);
                    }
                });
            } else if (child && typeof child === 'object' && child.type) {
                this.visitNode(child, nestingLevel);
            }
        });
    }

    isFunctionNode(node) {
        return node.type === 'FunctionDeclaration' ||
               node.type === 'FunctionExpression' ||
               node.type === 'ArrowFunctionExpression' ||
               node.type === 'ClassMethod' ||
               node.type === 'ObjectMethod';
    }

    visitFunction(node) {
        const savedComplexity = this.complexity;
        const savedNesting = this.nestingLevel;

        this.complexity = 0;
        this.nestingLevel = 0;

        // Visit function body
        if (node.body) {
            if (node.body.type === 'BlockStatement') {
                node.body.body.forEach(stmt => this.visitNode(stmt, 0));
            } else {
                // Arrow function with expression body
                this.visitNode(node.body, 0);
            }
        }

        // Get function name
        let name = 'anonymous';
        if (node.id && node.id.name) {
            name = node.id.name;
        } else if (node.key && node.key.name) {
            name = node.key.name;
        }

        this.functionMetrics.push({
            name: name,
            start_line: node.loc ? node.loc.start.line : 0,
            end_line: node.loc ? node.loc.end.line : 0,
            complexity: this.complexity,
            value: this.complexity
        });

        this.complexity = savedComplexity;
        this.nestingLevel = savedNesting;
    }

    calculate(ast) {
        this.visitNode(ast);

        const totalComplexity = this.functionMetrics.reduce((sum, f) => sum + f.complexity, 0);
        const functionCount = this.functionMetrics.length;
        const avgComplexity = functionCount > 0 ? totalComplexity / functionCount : 0;
        const maxComplexity = this.functionMetrics.reduce((max, f) => Math.max(max, f.complexity), 0);

        return {
            functions: this.functionMetrics,
            file_level: {
                total: totalComplexity,
                average: avgComplexity,
                max: maxComplexity,
                function_count: functionCount
            }
        };
    }
}

/**
 * Calculate Halstead metrics
 */
class HalsteadCalculator {
    constructor() {
        this.operators = new Map();
        this.operands = new Map();
        this.totalOperators = 0;
        this.totalOperands = 0;
        this.functionMetrics = [];
    }

    addOperator(op) {
        this.operators.set(op, (this.operators.get(op) || 0) + 1);
        this.totalOperators++;
    }

    addOperand(operand) {
        this.operands.set(operand, (this.operands.get(operand) || 0) + 1);
        this.totalOperands++;
    }

    isBuiltin(name) {
        const builtins = new Set([
            'console', 'undefined', 'null', 'true', 'false',
            'Array', 'Object', 'String', 'Number', 'Boolean',
            'Math', 'Date', 'JSON', 'Promise', 'Symbol',
            'parseInt', 'parseFloat', 'isNaN', 'isFinite',
        ]);
        return builtins.has(name);
    }

    visitNode(node) {
        if (!node) return;

        // Handle function declarations
        if (this.isFunctionNode(node)) {
            this.visitFunction(node);
            return;
        }

        switch (node.type) {
            case 'BinaryExpression':
            case 'LogicalExpression':
                this.addOperator(node.operator);
                break;

            case 'UnaryExpression':
            case 'UpdateExpression':
                this.addOperator(node.operator);
                break;

            case 'AssignmentExpression':
                this.addOperator(node.operator);
                break;

            case 'IfStatement':
                this.addOperator('if');
                break;

            case 'ForStatement':
                this.addOperator('for');
                break;

            case 'ForInStatement':
                this.addOperator('for-in');
                break;

            case 'ForOfStatement':
                this.addOperator('for-of');
                break;

            case 'WhileStatement':
                this.addOperator('while');
                break;

            case 'DoWhileStatement':
                this.addOperator('do-while');
                break;

            case 'SwitchStatement':
                this.addOperator('switch');
                break;

            case 'SwitchCase':
                this.addOperator(node.test ? 'case' : 'default');
                break;

            case 'TryStatement':
                this.addOperator('try');
                break;

            case 'CatchClause':
                this.addOperator('catch');
                break;

            case 'ThrowStatement':
                this.addOperator('throw');
                break;

            case 'ReturnStatement':
                this.addOperator('return');
                break;

            case 'BreakStatement':
                this.addOperator('break');
                break;

            case 'ContinueStatement':
                this.addOperator('continue');
                break;

            case 'CallExpression':
            case 'NewExpression':
                this.addOperator('()');
                break;

            case 'MemberExpression':
                this.addOperator(node.computed ? '[]' : '.');
                break;

            case 'ConditionalExpression':
                this.addOperator('?:');
                break;

            case 'Identifier':
                if (!this.isBuiltin(node.name)) {
                    this.addOperand(node.name);
                }
                break;

            case 'Literal':
            case 'StringLiteral':
            case 'NumericLiteral':
            case 'BooleanLiteral':
                this.addOperand(String(node.value));
                break;
        }

        // Visit children
        this.visitChildren(node);
    }

    visitChildren(node) {
        if (!node || typeof node !== 'object') return;

        Object.keys(node).forEach(key => {
            const child = node[key];
            if (Array.isArray(child)) {
                child.forEach(item => {
                    if (item && typeof item === 'object' && item.type) {
                        this.visitNode(item);
                    }
                });
            } else if (child && typeof child === 'object' && child.type) {
                this.visitNode(child);
            }
        });
    }

    isFunctionNode(node) {
        return node.type === 'FunctionDeclaration' ||
               node.type === 'FunctionExpression' ||
               node.type === 'ArrowFunctionExpression' ||
               node.type === 'ClassMethod' ||
               node.type === 'ObjectMethod';
    }

    visitFunction(node) {
        const savedOperators = this.operators;
        const savedOperands = this.operands;
        const savedTotalOps = this.totalOperators;
        const savedTotalOpds = this.totalOperands;

        this.operators = new Map();
        this.operands = new Map();
        this.totalOperators = 0;
        this.totalOperands = 0;

        // Add function declaration as operator
        this.addOperator('function');

        // Add parameters as operands
        if (node.params) {
            node.params.forEach(param => {
                if (param.type === 'Identifier') {
                    this.addOperand(param.name);
                }
            });
        }

        // Visit function body
        if (node.body) {
            this.visitNode(node.body);
        }

        // Calculate metrics
        const metrics = this.calculateMetrics();

        // Get function name
        let name = 'anonymous';
        if (node.id && node.id.name) {
            name = node.id.name;
        } else if (node.key && node.key.name) {
            name = node.key.name;
        }

        this.functionMetrics.push({
            name: name,
            start_line: node.loc ? node.loc.start.line : 0,
            end_line: node.loc ? node.loc.end.line : 0,
            value: metrics.effort
        });

        this.operators = savedOperators;
        this.operands = savedOperands;
        this.totalOperators = savedTotalOps;
        this.totalOperands = savedTotalOpds;
    }

    calculateMetrics() {
        const n1 = this.operators.size;  // Distinct operators
        const n2 = this.operands.size;   // Distinct operands
        const N1 = this.totalOperators;   // Total operators
        const N2 = this.totalOperands;    // Total operands

        const vocabulary = n1 + n2;
        const length = N1 + N2;

        let volume = 0;
        if (vocabulary > 0) {
            volume = length * Math.log2(vocabulary);
        }

        let difficulty = 0;
        if (n2 > 0) {
            difficulty = (n1 / 2.0) * (N2 / n2);
        }

        const effort = difficulty * volume;

        return {
            distinct_operators: n1,
            distinct_operands: n2,
            total_operators: N1,
            total_operands: N2,
            vocabulary,
            length,
            volume,
            difficulty,
            effort
        };
    }

    calculate(ast) {
        this.visitNode(ast);

        const totalEffort = this.functionMetrics.reduce((sum, f) => sum + f.value, 0);
        const functionCount = this.functionMetrics.length;
        const avgEffort = functionCount > 0 ? totalEffort / functionCount : 0;
        const maxEffort = this.functionMetrics.reduce((max, f) => Math.max(max, f.value), 0);

        return {
            functions: this.functionMetrics,
            file_level: {
                total_effort: totalEffort,
                average_effort: avgEffort,
                max_effort: maxEffort,
                function_count: functionCount
            }
        };
    }
}

/**
 * Parse a JavaScript/TypeScript file
 */
function parseFile(filePath) {
    const source = fs.readFileSync(filePath, 'utf-8');
    const isTypeScript = filePath.endsWith('.ts') || filePath.endsWith('.tsx');

    const options = {
        sourceType: 'module',
        locations: true,
        plugins: [
            'jsx',
            'classProperties',
            'decorators-legacy',
            'exportDefaultFrom',
            'exportNamespaceFrom',
            'dynamicImport',
            'optionalChaining',
            'nullishCoalescingOperator',
            'optionalCatchBinding'
        ]
    };

    if (isTypeScript) {
        options.plugins.push('typescript');
    }

    return parser.parse(source, options);
}

/**
 * Calculate cognitive complexity for a file
 */
function calculateCognitiveComplexity(filePath) {
    try {
        const ast = parseFile(filePath);
        const calculator = new CognitiveComplexityCalculator();
        return calculator.calculate(ast);
    } catch (error) {
        return {
            error: error.message,
            functions: [],
            file_level: {}
        };
    }
}

/**
 * Calculate Halstead metrics for a file
 */
function calculateHalstead(filePath) {
    try {
        const ast = parseFile(filePath);
        const calculator = new HalsteadCalculator();
        return calculator.calculate(ast);
    } catch (error) {
        return {
            error: error.message,
            functions: [],
            file_level: {}
        };
    }
}

/**
 * Main entry point
 */
function main() {
    if (process.argv.length < 4) {
        console.log(JSON.stringify({
            error: 'Usage: node js_metrics.js <metric_type> <file_path>'
        }));
        process.exit(1);
    }

    const metricType = process.argv[2];
    const filePath = process.argv[3];

    let result;
    if (metricType === 'cognitive-complexity') {
        result = calculateCognitiveComplexity(filePath);
    } else if (metricType === 'halstead') {
        result = calculateHalstead(filePath);
    } else {
        result = { error: `Unknown metric type: ${metricType}` };
    }

    console.log(JSON.stringify(result, null, 2));
}

main();
