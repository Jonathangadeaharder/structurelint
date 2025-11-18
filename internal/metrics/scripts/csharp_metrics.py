#!/usr/bin/env python3
"""
C# metrics calculator for structurelint.
Calculates cognitive complexity and Halstead metrics for C# code using tree-sitter.
"""

import sys
import json
import math
from collections import defaultdict
from typing import Dict, Any, List

try:
    from tree_sitter import Language, Parser, Node
    import tree_sitter_c_sharp as tscsharp
except ImportError:
    print(json.dumps({
        'error': 'tree-sitter or tree-sitter-c-sharp not installed. Install with: pip install tree-sitter tree-sitter-c-sharp'
    }))
    sys.exit(1)


class CognitiveComplexityCalculator:
    """Calculates cognitive complexity for C# code."""

    def __init__(self, source_code: bytes):
        self.source = source_code
        self.parser = Parser()
        self.parser.set_language(Language(tscsharp.language()))
        self.tree = self.parser.parse(source_code)
        self.function_metrics = []

    def calculate(self) -> List[Dict[str, Any]]:
        """Calculate cognitive complexity for all methods."""
        root = self.tree.root_node
        self._visit_methods(root)
        return self.function_metrics

    def _visit_methods(self, node: Node):
        """Recursively find and analyze all method declarations."""
        if node.type in ['method_declaration', 'constructor_declaration',
                        'local_function_statement']:
            self._analyze_method(node)

        for child in node.children:
            self._visit_methods(child)

    def _analyze_method(self, node: Node):
        """Analyze a single method."""
        # Get method name
        method_name = self._get_method_name(node)

        # Calculate complexity
        complexity = self._calculate_complexity(node, nesting_level=0)

        # Get line numbers
        start_line = node.start_point[0] + 1
        end_line = node.end_point[0] + 1

        self.function_metrics.append({
            'name': method_name,
            'start_line': start_line,
            'end_line': end_line,
            'complexity': complexity,
            'value': float(complexity)
        })

    def _get_method_name(self, node: Node) -> str:
        """Extract method name from method_declaration node."""
        for child in node.children:
            if child.type == 'identifier':
                return self.source[child.start_byte:child.end_byte].decode('utf-8')
        return 'unknown'

    def _calculate_complexity(self, node: Node, nesting_level: int) -> int:
        """Calculate cognitive complexity recursively."""
        complexity = 0

        # Control structures that add complexity
        if node.type in ['if_statement', 'while_statement', 'for_statement',
                        'foreach_statement', 'do_statement']:
            complexity += 1 + nesting_level
            nesting_level += 1

        # Switch statements and expressions
        elif node.type in ['switch_statement', 'switch_expression']:
            complexity += 1 + nesting_level
            nesting_level += 1

        # Catch blocks
        elif node.type == 'catch_clause':
            complexity += 1 + nesting_level
            nesting_level += 1

        # Conditional expressions (ternary)
        elif node.type == 'conditional_expression':
            complexity += 1 + nesting_level

        # LINQ query expressions
        elif node.type == 'query_expression':
            complexity += 1 + nesting_level
            nesting_level += 1

        # Null-coalescing operators
        elif node.type == 'binary_expression':
            operator = self._get_operator(node)
            if operator in ['??', '&&', '||']:
                complexity += 1

        # Lambda expressions increase nesting
        elif node.type in ['lambda_expression', 'anonymous_method_expression']:
            nesting_level += 1

        # Pattern matching (switch expressions)
        elif node.type == 'switch_expression_arm':
            complexity += 1

        # Recursively process children
        for child in node.children:
            complexity += self._calculate_complexity(child, nesting_level)

        return complexity

    def _get_operator(self, node: Node) -> str:
        """Extract operator from binary expression."""
        for child in node.children:
            text = self.source[child.start_byte:child.end_byte].decode('utf-8')
            if text in ['&&', '||', '??', '?.']:
                return text
        return ''


class HalsteadCalculator:
    """Calculates Halstead metrics for C# code."""

    def __init__(self, source_code: bytes):
        self.source = source_code
        self.parser = Parser()
        self.parser.set_language(Language(tscsharp.language()))
        self.tree = self.parser.parse(source_code)
        self.function_metrics = []

    def calculate(self) -> List[Dict[str, Any]]:
        """Calculate Halstead metrics for all methods."""
        root = self.tree.root_node
        self._visit_methods(root)
        return self.function_metrics

    def _visit_methods(self, node: Node):
        """Recursively find and analyze all method declarations."""
        if node.type in ['method_declaration', 'constructor_declaration',
                        'local_function_statement']:
            self._analyze_method(node)

        for child in node.children:
            self._visit_methods(child)

    def _analyze_method(self, node: Node):
        """Analyze Halstead metrics for a single method."""
        method_name = self._get_method_name(node)

        # Count operators and operands
        operators = defaultdict(int)
        operands = defaultdict(int)

        self._count_operators_operands(node, operators, operands)

        # Calculate metrics
        metrics = self._calculate_halstead_metrics(operators, operands)

        # Get line numbers
        start_line = node.start_point[0] + 1
        end_line = node.end_point[0] + 1

        self.function_metrics.append({
            'name': method_name,
            'start_line': start_line,
            'end_line': end_line,
            'value': metrics['effort']
        })

    def _get_method_name(self, node: Node) -> str:
        """Extract method name from method_declaration node."""
        for child in node.children:
            if child.type == 'identifier':
                return self.source[child.start_byte:child.end_byte].decode('utf-8')
        return 'unknown'

    def _count_operators_operands(self, node: Node, operators: Dict, operands: Dict):
        """Recursively count operators and operands."""
        # Operators
        if node.type in ['binary_expression', 'prefix_unary_expression', 'postfix_unary_expression',
                        'assignment_expression', 'conditional_expression']:
            op = self._extract_operator(node)
            if op:
                operators[op] += 1

        # Control flow operators
        elif node.type in ['if_statement', 'while_statement', 'for_statement',
                          'foreach_statement', 'do_statement', 'switch_statement',
                          'try_statement', 'query_expression']:
            operators[node.type.replace('_statement', '').replace('_expression', '')] += 1

        # Method calls/invocations
        elif node.type == 'invocation_expression':
            operators['()'] += 1

        # Array/element access
        elif node.type == 'element_access_expression':
            operators['[]'] += 1

        # Member access
        elif node.type == 'member_access_expression':
            operators['.'] += 1

        # Null-conditional access
        elif node.type == 'conditional_access_expression':
            operators['?.'] += 1

        # Operands (identifiers, literals)
        elif node.type == 'identifier':
            name = self.source[node.start_byte:node.end_byte].decode('utf-8')
            if not self._is_keyword(name):
                operands[name] += 1

        elif node.type in ['integer_literal', 'real_literal', 'character_literal',
                          'string_literal', 'verbatim_string_literal', 'interpolated_string_expression',
                          'true_literal', 'false_literal', 'null_literal']:
            value = self.source[node.start_byte:node.end_byte].decode('utf-8')
            # Truncate long strings for counting
            if len(value) > 50:
                value = value[:50] + '...'
            operands[value] += 1

        # Recurse
        for child in node.children:
            self._count_operators_operands(child, operators, operands)

    def _extract_operator(self, node: Node) -> str:
        """Extract operator from expression node."""
        for child in node.children:
            text = self.source[child.start_byte:child.end_byte].decode('utf-8')
            if text in ['+', '-', '*', '/', '%', '=', '==', '!=', '<', '>', '<=', '>=',
                       '&&', '||', '!', '&', '|', '^', '~', '<<', '>>',
                       '+=', '-=', '*=', '/=', '%=', '&=', '|=', '^=', '<<=', '>>=',
                       '++', '--', '?', ':', '??', '??=', '?.', '=>']:
                return text
        return ''

    def _is_keyword(self, name: str) -> bool:
        """Check if a name is a C# keyword."""
        keywords = {
            'abstract', 'as', 'base', 'bool', 'break', 'byte', 'case', 'catch',
            'char', 'checked', 'class', 'const', 'continue', 'decimal', 'default',
            'delegate', 'do', 'double', 'else', 'enum', 'event', 'explicit', 'extern',
            'false', 'finally', 'fixed', 'float', 'for', 'foreach', 'goto', 'if',
            'implicit', 'in', 'int', 'interface', 'internal', 'is', 'lock', 'long',
            'namespace', 'new', 'null', 'object', 'operator', 'out', 'override',
            'params', 'private', 'protected', 'public', 'readonly', 'ref', 'return',
            'sbyte', 'sealed', 'short', 'sizeof', 'stackalloc', 'static', 'string',
            'struct', 'switch', 'this', 'throw', 'true', 'try', 'typeof', 'uint',
            'ulong', 'unchecked', 'unsafe', 'ushort', 'using', 'virtual', 'void',
            'volatile', 'while', 'add', 'alias', 'ascending', 'async', 'await',
            'by', 'descending', 'dynamic', 'equals', 'from', 'get', 'global',
            'group', 'into', 'join', 'let', 'nameof', 'on', 'orderby', 'partial',
            'remove', 'select', 'set', 'value', 'var', 'when', 'where', 'yield'
        }
        return name in keywords

    def _calculate_halstead_metrics(self, operators: Dict, operands: Dict) -> Dict[str, float]:
        """Calculate Halstead metrics from operator and operand counts."""
        n1 = len(operators)  # Distinct operators
        n2 = len(operands)   # Distinct operands
        N1 = sum(operators.values())  # Total operators
        N2 = sum(operands.values())   # Total operands

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
    """Calculate cognitive complexity for a C# file."""
    try:
        with open(file_path, 'rb') as f:
            source_code = f.read()

        calculator = CognitiveComplexityCalculator(source_code)
        functions = calculator.calculate()

        # Calculate file-level metrics
        total_complexity = sum(f['complexity'] for f in functions)
        function_count = len(functions)
        avg_complexity = total_complexity / function_count if function_count > 0 else 0
        max_complexity = max((f['complexity'] for f in functions), default=0)

        return {
            'functions': functions,
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
    """Calculate Halstead metrics for a C# file."""
    try:
        with open(file_path, 'rb') as f:
            source_code = f.read()

        calculator = HalsteadCalculator(source_code)
        functions = calculator.calculate()

        # Calculate file-level metrics
        total_effort = sum(f['value'] for f in functions)
        function_count = len(functions)
        avg_effort = total_effort / function_count if function_count > 0 else 0
        max_effort = max((f['value'] for f in functions), default=0)

        return {
            'functions': functions,
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
            'error': 'Usage: csharp_metrics.py <metric_type> <file_path>'
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

    # Propagate failures so callers can detect them
    if isinstance(result, dict) and result.get('error'):
        sys.exit(1)


if __name__ == '__main__':
    main()
