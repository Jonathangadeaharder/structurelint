#!/usr/bin/env python3
"""
Python metrics calculator for structurelint.
Calculates cognitive complexity and Halstead metrics for Python code.
"""

import ast
import sys
import json
import math
from collections import defaultdict
from typing import Dict, Any


class CognitiveComplexityVisitor(ast.NodeVisitor):
    """Calculates cognitive complexity using the same rules as Go implementation."""

    def __init__(self):
        self.complexity = 0
        self.nesting_level = 0
        self.function_metrics = []
        self.current_function = None

    def visit_FunctionDef(self, node):
        """Visit function definition and calculate its complexity."""
        # Save previous function context
        prev_function = self.current_function
        prev_complexity = self.complexity
        prev_nesting = self.nesting_level

        # Start new function
        self.current_function = node.name
        self.complexity = 0
        self.nesting_level = 0

        # Visit function body
        for stmt in node.body:
            self.visit(stmt)

        # Record metrics for this function
        self.function_metrics.append({
            'name': node.name,
            'start_line': node.lineno,
            'end_line': node.end_lineno or node.lineno,
            'complexity': self.complexity,
            'value': float(self.complexity)
        })

        # Restore previous context
        self.current_function = prev_function
        self.complexity = prev_complexity
        self.nesting_level = prev_nesting

    visit_AsyncFunctionDef = visit_FunctionDef

    def visit_If(self, node):
        """If statements add 1 + nesting_level to complexity."""
        self.complexity += 1 + self.nesting_level

        # Visit the condition to catch boolean operators
        if node.test:
            self.visit(node.test)

        self.nesting_level += 1

        for stmt in node.body:
            self.visit(stmt)

        self.nesting_level -= 1

        # Handle elif/else chains
        if node.orelse:
            if len(node.orelse) == 1 and isinstance(node.orelse[0], ast.If):
                # This is an elif - don't increase nesting
                self.visit(node.orelse[0])
            else:
                # This is an else block
                self.nesting_level += 1
                for stmt in node.orelse:
                    self.visit(stmt)
                self.nesting_level -= 1

    def visit_For(self, node):
        """For loops add 1 + nesting_level to complexity."""
        self.complexity += 1 + self.nesting_level

        # Visit the iterator and target
        if node.target:
            self.visit(node.target)
        if node.iter:
            self.visit(node.iter)

        self.nesting_level += 1

        for stmt in node.body:
            self.visit(stmt)

        self.nesting_level -= 1

        if node.orelse:
            for stmt in node.orelse:
                self.visit(stmt)

    visit_AsyncFor = visit_For

    def visit_While(self, node):
        """While loops add 1 + nesting_level to complexity."""
        self.complexity += 1 + self.nesting_level

        # Visit the condition to catch boolean operators
        if node.test:
            self.visit(node.test)

        self.nesting_level += 1

        for stmt in node.body:
            self.visit(stmt)

        self.nesting_level -= 1

        if node.orelse:
            for stmt in node.orelse:
                self.visit(stmt)

    def visit_Try(self, node):
        """Try statements add 1 + nesting_level to complexity."""
        self.complexity += 1 + self.nesting_level
        self.nesting_level += 1

        for stmt in node.body:
            self.visit(stmt)

        self.nesting_level -= 1

        # Each except handler adds complexity
        for handler in node.handlers:
            self.complexity += 1
            self.nesting_level += 1
            for stmt in handler.body:
                self.visit(stmt)
            self.nesting_level -= 1

        if node.orelse:
            for stmt in node.orelse:
                self.visit(stmt)

        if node.finalbody:
            for stmt in node.finalbody:
                self.visit(stmt)

    visit_TryStar = visit_Try

    def visit_With(self, node):
        """With statements add 1 + nesting_level to complexity."""
        self.complexity += 1 + self.nesting_level
        self.nesting_level += 1

        for stmt in node.body:
            self.visit(stmt)

        self.nesting_level -= 1

    visit_AsyncWith = visit_With

    def visit_Break(self, node):
        """Break statements add 1 + nesting_level to complexity."""
        self.complexity += 1 + self.nesting_level

    def visit_Continue(self, node):
        """Continue statements add 1 + nesting_level to complexity."""
        self.complexity += 1 + self.nesting_level

    def visit_BoolOp(self, node):
        """Boolean operators (and/or) add to complexity."""
        if isinstance(node.op, (ast.And, ast.Or)):
            # Each additional operand adds 1
            self.complexity += len(node.values) - 1
        self.generic_visit(node)


class HalsteadVisitor(ast.NodeVisitor):
    """Calculates Halstead metrics for Python code."""

    def __init__(self):
        self.operators = defaultdict(int)
        self.operands = defaultdict(int)
        self.total_operators = 0
        self.total_operands = 0
        self.function_metrics = []
        self.current_function = None

    def visit_FunctionDef(self, node):
        """Visit function definition and calculate Halstead metrics."""
        # Save previous function context
        prev_function = self.current_function
        prev_operators = self.operators
        prev_operands = self.operands
        prev_total_ops = self.total_operators
        prev_total_opds = self.total_operands

        # Start new function
        self.current_function = node.name
        self.operators = defaultdict(int)
        self.operands = defaultdict(int)
        self.total_operators = 0
        self.total_operands = 0

        # Add function definition as operator
        self.add_operator('def')

        # Add parameters as operands
        for arg in node.args.args:
            self.add_operand(arg.arg)

        # Visit function body
        for stmt in node.body:
            self.visit(stmt)

        # Calculate metrics
        metrics = self.calculate_metrics()
        self.function_metrics.append({
            'name': node.name,
            'start_line': node.lineno,
            'end_line': node.end_lineno or node.lineno,
            'value': metrics['effort']
        })

        # Restore previous context
        self.current_function = prev_function
        self.operators = prev_operators
        self.operands = prev_operands
        self.total_operators = prev_total_ops
        self.total_operands = prev_total_opds

    visit_AsyncFunctionDef = visit_FunctionDef

    def add_operator(self, op: str):
        """Add an operator to the count."""
        self.operators[op] += 1
        self.total_operators += 1

    def add_operand(self, operand: str):
        """Add an operand to the count."""
        self.operands[operand] += 1
        self.total_operands += 1

    def visit_BinOp(self, node):
        """Binary operations."""
        self.add_operator(node.op.__class__.__name__)
        self.generic_visit(node)

    def visit_UnaryOp(self, node):
        """Unary operations."""
        self.add_operator(node.op.__class__.__name__)
        self.generic_visit(node)

    def visit_Compare(self, node):
        """Comparison operations."""
        for op in node.ops:
            self.add_operator(op.__class__.__name__)
        self.generic_visit(node)

    def visit_BoolOp(self, node):
        """Boolean operations."""
        self.add_operator(node.op.__class__.__name__)
        self.generic_visit(node)

    def visit_Assign(self, node):
        """Assignment."""
        self.add_operator('=')
        self.generic_visit(node)

    def visit_AugAssign(self, node):
        """Augmented assignment."""
        self.add_operator(node.op.__class__.__name__ + '=')
        self.generic_visit(node)

    def visit_AnnAssign(self, node):
        """Annotated assignment."""
        self.add_operator('=')
        self.generic_visit(node)

    def visit_If(self, node):
        """If statement."""
        self.add_operator('if')
        self.generic_visit(node)

    def visit_For(self, node):
        """For loop."""
        self.add_operator('for')
        self.generic_visit(node)

    visit_AsyncFor = visit_For

    def visit_While(self, node):
        """While loop."""
        self.add_operator('while')
        self.generic_visit(node)

    def visit_Try(self, node):
        """Try statement."""
        self.add_operator('try')
        for handler in node.handlers:
            self.add_operator('except')
        self.generic_visit(node)

    visit_TryStar = visit_Try

    def visit_With(self, node):
        """With statement."""
        self.add_operator('with')
        self.generic_visit(node)

    visit_AsyncWith = visit_With

    def visit_Return(self, node):
        """Return statement."""
        self.add_operator('return')
        self.generic_visit(node)

    def visit_Break(self, node):
        """Break statement."""
        self.add_operator('break')

    def visit_Continue(self, node):
        """Continue statement."""
        self.add_operator('continue')

    def visit_Call(self, node):
        """Function call."""
        self.add_operator('()')
        self.generic_visit(node)

    def visit_Subscript(self, node):
        """Subscript access."""
        self.add_operator('[]')
        self.generic_visit(node)

    def visit_Name(self, node):
        """Variable names."""
        # Don't count keywords or builtins
        if not self.is_builtin(node.id):
            self.add_operand(node.id)

    def visit_Constant(self, node):
        """Literal constants."""
        self.add_operand(str(node.value))

    def is_builtin(self, name: str) -> bool:
        """Check if a name is a built-in."""
        builtins = {
            'True', 'False', 'None', 'and', 'or', 'not', 'is', 'in',
            'print', 'len', 'range', 'str', 'int', 'float', 'bool',
            'list', 'dict', 'set', 'tuple', 'type', 'object',
            'abs', 'all', 'any', 'sum', 'min', 'max', 'sorted',
            'open', 'input', 'isinstance', 'issubclass',
        }
        return name in builtins

    def calculate_metrics(self) -> Dict[str, float]:
        """Calculate Halstead metrics."""
        n1 = len(self.operators)  # Distinct operators
        n2 = len(self.operands)   # Distinct operands
        N1 = self.total_operators  # Total operators
        N2 = self.total_operands   # Total operands

        vocabulary = n1 + n2
        length = N1 + N2

        volume = 0.0
        if vocabulary > 0:
            volume = float(length) * math.log2(vocabulary)

        difficulty = 0.0
        if n2 > 0:
            difficulty = (float(n1) / 2.0) * (float(N2) / float(n2))

        effort = difficulty * volume

        return {
            'distinct_operators': n1,
            'distinct_operands': n2,
            'total_operators': N1,
            'total_operands': N2,
            'vocabulary': vocabulary,
            'length': length,
            'volume': volume,
            'difficulty': difficulty,
            'effort': effort
        }


def calculate_cognitive_complexity(file_path: str) -> Dict[str, Any]:
    """Calculate cognitive complexity for a Python file."""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            source = f.read()

        tree = ast.parse(source, filename=file_path)
        visitor = CognitiveComplexityVisitor()
        visitor.visit(tree)

        # Calculate file-level metrics
        total_complexity = sum(f['complexity'] for f in visitor.function_metrics)
        function_count = len(visitor.function_metrics)
        avg_complexity = total_complexity / function_count if function_count > 0 else 0
        max_complexity = max((f['complexity'] for f in visitor.function_metrics), default=0)

        return {
            'functions': visitor.function_metrics,
            'file_level': {
                'total': float(total_complexity),
                'average': float(avg_complexity),
                'max': float(max_complexity),
                'function_count': float(function_count)
            }
        }
    except Exception as e:
        return {
            'error': str(e),
            'functions': [],
            'file_level': {}
        }


def calculate_halstead(file_path: str) -> Dict[str, Any]:
    """Calculate Halstead metrics for a Python file."""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            source = f.read()

        tree = ast.parse(source, filename=file_path)
        visitor = HalsteadVisitor()
        visitor.visit(tree)

        # Calculate file-level metrics
        total_effort = sum(f['value'] for f in visitor.function_metrics)
        function_count = len(visitor.function_metrics)
        avg_effort = total_effort / function_count if function_count > 0 else 0
        max_effort = max((f['value'] for f in visitor.function_metrics), default=0)

        return {
            'functions': visitor.function_metrics,
            'file_level': {
                'total_effort': float(total_effort),
                'average_effort': float(avg_effort),
                'max_effort': float(max_effort),
                'function_count': float(function_count)
            }
        }
    except Exception as e:
        return {
            'error': str(e),
            'functions': [],
            'file_level': {}
        }


def main():
    """Main entry point."""
    if len(sys.argv) < 3:
        print(json.dumps({
            'error': 'Usage: python_metrics.py <metric_type> <file_path>'
        }))
        sys.exit(1)

    metric_type = sys.argv[1]
    file_path = sys.argv[2]

    if metric_type == 'cognitive-complexity':
        result = calculate_cognitive_complexity(file_path)
    elif metric_type == 'halstead':
        result = calculate_halstead(file_path)
    else:
        result = {'error': f'Unknown metric type: {metric_type}'}

    print(json.dumps(result, indent=2))

    # Propagate failures so callers (e.g. the Go MultiLanguageAnalyzer) can detect them
    if isinstance(result, dict) and result.get('error'):
        sys.exit(1)


if __name__ == '__main__':
    main()
