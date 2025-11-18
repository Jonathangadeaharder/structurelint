#!/usr/bin/env python3
"""
C++ metrics calculator for structurelint.
Calculates cognitive complexity and Halstead metrics for C++ code using tree-sitter.
"""

import sys
import json
import math
from collections import defaultdict
from typing import Dict, Any, List

try:
    from tree_sitter import Language, Parser, Node
    import tree_sitter_cpp as tscpp
except ImportError:
    print(json.dumps({
        'error': 'tree-sitter or tree-sitter-cpp not installed. Install with: pip install tree-sitter tree-sitter-cpp'
    }))
    sys.exit(1)


class CognitiveComplexityCalculator:
    """Calculates cognitive complexity for C++ code."""

    def __init__(self, source_code: bytes):
        self.source = source_code
        self.parser = Parser()
        self.parser.set_language(Language(tscpp.language()))
        self.tree = self.parser.parse(source_code)
        self.function_metrics = []

    def calculate(self) -> List[Dict[str, Any]]:
        """Calculate cognitive complexity for all functions."""
        root = self.tree.root_node
        self._visit_functions(root)
        return self.function_metrics

    def _visit_functions(self, node: Node):
        """Recursively find and analyze all function definitions."""
        if node.type == 'function_definition':
            self._analyze_function(node)

        for child in node.children:
            self._visit_functions(child)

    def _analyze_function(self, node: Node):
        """Analyze a single function."""
        # Get function name
        function_name = self._get_function_name(node)

        # Calculate complexity
        complexity = self._calculate_complexity(node, nesting_level=0)

        # Get line numbers
        start_line = node.start_point[0] + 1
        end_line = node.end_point[0] + 1

        self.function_metrics.append({
            'name': function_name,
            'start_line': start_line,
            'end_line': end_line,
            'complexity': complexity,
            'value': float(complexity)
        })

    def _get_function_name(self, node: Node) -> str:
        """Extract function name from function_definition node."""
        # Look for function_declarator
        for child in node.children:
            if child.type == 'function_declarator':
                return self._extract_declarator_name(child)
        return 'unknown'

    def _extract_declarator_name(self, node: Node) -> str:
        """Extract name from declarator."""
        for child in node.children:
            if child.type == 'identifier':
                return self.source[child.start_byte:child.end_byte].decode('utf-8')
            elif child.type == 'field_identifier':
                return self.source[child.start_byte:child.end_byte].decode('utf-8')
            elif child.type in ['qualified_identifier', 'scoped_identifier']:
                # For scoped names like MyClass::method
                parts = []
                self._extract_scoped_name(child, parts)
                return '::'.join(parts)
        return 'unknown'

    def _extract_scoped_name(self, node: Node, parts: List[str]):
        """Recursively extract scoped identifier parts."""
        for child in node.children:
            if child.type == 'identifier' or child.type == 'field_identifier':
                parts.append(self.source[child.start_byte:child.end_byte].decode('utf-8'))
            elif child.type in ['namespace_identifier', 'type_identifier']:
                parts.append(self.source[child.start_byte:child.end_byte].decode('utf-8'))
            else:
                self._extract_scoped_name(child, parts)

    def _calculate_complexity(self, node: Node, nesting_level: int) -> int:
        """Calculate cognitive complexity recursively."""
        complexity = 0

        # Control structures that add complexity
        if node.type in ['if_statement', 'while_statement', 'for_statement',
                        'for_range_loop', 'do_statement']:
            complexity += 1 + nesting_level
            nesting_level += 1

        # Switch statements
        elif node.type == 'switch_statement':
            complexity += 1 + nesting_level
            nesting_level += 1

        # Catch blocks
        elif node.type == 'catch_clause':
            complexity += 1 + nesting_level
            nesting_level += 1

        # Conditional expressions (ternary)
        elif node.type == 'conditional_expression':
            complexity += 1 + nesting_level

        # Goto statements
        elif node.type == 'goto_statement':
            complexity += 1

        # Lambda expressions increase nesting
        elif node.type == 'lambda_expression':
            nesting_level += 1

        # Boolean operators in conditions
        elif node.type == 'binary_expression':
            operator = self._get_operator(node)
            if operator in ['&&', '||']:
                complexity += 1

        # Recursively process children
        for child in node.children:
            complexity += self._calculate_complexity(child, nesting_level)

        return complexity

    def _get_operator(self, node: Node) -> str:
        """Extract operator from binary expression."""
        for child in node.children:
            text = self.source[child.start_byte:child.end_byte].decode('utf-8')
            if text in ['&&', '||', '&', '|', 'and', 'or']:
                return text
        return ''


class HalsteadCalculator:
    """Calculates Halstead metrics for C++ code."""

    def __init__(self, source_code: bytes):
        self.source = source_code
        self.parser = Parser()
        self.parser.set_language(Language(tscpp.language()))
        self.tree = self.parser.parse(source_code)
        self.function_metrics = []

    def calculate(self) -> List[Dict[str, Any]]:
        """Calculate Halstead metrics for all functions."""
        root = self.tree.root_node
        self._visit_functions(root)
        return self.function_metrics

    def _visit_functions(self, node: Node):
        """Recursively find and analyze all function definitions."""
        if node.type == 'function_definition':
            self._analyze_function(node)

        for child in node.children:
            self._visit_functions(child)

    def _analyze_function(self, node: Node):
        """Analyze Halstead metrics for a single function."""
        function_name = self._get_function_name(node)

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
            'name': function_name,
            'start_line': start_line,
            'end_line': end_line,
            'value': metrics['effort']
        })

    def _get_function_name(self, node: Node) -> str:
        """Extract function name from function_definition node."""
        # Look for function_declarator
        for child in node.children:
            if child.type == 'function_declarator':
                return self._extract_declarator_name(child)
        return 'unknown'

    def _extract_declarator_name(self, node: Node) -> str:
        """Extract name from declarator."""
        for child in node.children:
            if child.type == 'identifier':
                return self.source[child.start_byte:child.end_byte].decode('utf-8')
            elif child.type == 'field_identifier':
                return self.source[child.start_byte:child.end_byte].decode('utf-8')
            elif child.type in ['qualified_identifier', 'scoped_identifier']:
                # For scoped names like MyClass::method
                parts = []
                self._extract_scoped_name(child, parts)
                return '::'.join(parts)
        return 'unknown'

    def _extract_scoped_name(self, node: Node, parts: List[str]):
        """Recursively extract scoped identifier parts."""
        for child in node.children:
            if child.type == 'identifier' or child.type == 'field_identifier':
                parts.append(self.source[child.start_byte:child.end_byte].decode('utf-8'))
            elif child.type in ['namespace_identifier', 'type_identifier']:
                parts.append(self.source[child.start_byte:child.end_byte].decode('utf-8'))
            else:
                self._extract_scoped_name(child, parts)

    def _count_operators_operands(self, node: Node, operators: Dict, operands: Dict):
        """Recursively count operators and operands."""
        # Operators
        if node.type in ['binary_expression', 'unary_expression', 'update_expression',
                        'assignment_expression', 'conditional_expression']:
            op = self._extract_operator(node)
            if op:
                operators[op] += 1

        # Control flow operators
        elif node.type in ['if_statement', 'while_statement', 'for_statement',
                          'do_statement', 'switch_statement', 'try_statement',
                          'for_range_loop']:
            operators[node.type.replace('_statement', '').replace('_loop', '')] += 1

        # Function calls
        elif node.type == 'call_expression':
            operators['()'] += 1

        # Array/subscript access
        elif node.type == 'subscript_expression':
            operators['[]'] += 1

        # Pointer/member access
        elif node.type == 'field_expression':
            op = self._get_field_operator(node)
            if op:
                operators[op] += 1

        # Operands (identifiers, literals)
        elif node.type == 'identifier':
            name = self.source[node.start_byte:node.end_byte].decode('utf-8')
            if not self._is_keyword(name):
                operands[name] += 1

        elif node.type in ['number_literal', 'string_literal', 'char_literal',
                          'true', 'false', 'null', 'nullptr']:
            value = self.source[node.start_byte:node.end_byte].decode('utf-8')
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
                       '++', '--', '?', ':', '->', '.', '::', 'and', 'or', 'not']:
                return text
        return ''

    def _get_field_operator(self, node: Node) -> str:
        """Get field access operator (. or ->)."""
        for child in node.children:
            text = self.source[child.start_byte:child.end_byte].decode('utf-8')
            if text in ['.', '->']:
                return text
        return ''

    def _is_keyword(self, name: str) -> bool:
        """Check if a name is a C++ keyword."""
        keywords = {
            'alignas', 'alignof', 'and', 'and_eq', 'asm', 'auto', 'bitand', 'bitor',
            'bool', 'break', 'case', 'catch', 'char', 'char8_t', 'char16_t', 'char32_t',
            'class', 'compl', 'concept', 'const', 'consteval', 'constexpr', 'constinit',
            'const_cast', 'continue', 'co_await', 'co_return', 'co_yield', 'decltype',
            'default', 'delete', 'do', 'double', 'dynamic_cast', 'else', 'enum', 'explicit',
            'export', 'extern', 'false', 'float', 'for', 'friend', 'goto', 'if', 'inline',
            'int', 'long', 'mutable', 'namespace', 'new', 'noexcept', 'not', 'not_eq',
            'nullptr', 'operator', 'or', 'or_eq', 'private', 'protected', 'public',
            'register', 'reinterpret_cast', 'requires', 'return', 'short', 'signed',
            'sizeof', 'static', 'static_assert', 'static_cast', 'struct', 'switch',
            'template', 'this', 'thread_local', 'throw', 'true', 'try', 'typedef',
            'typeid', 'typename', 'union', 'unsigned', 'using', 'virtual', 'void',
            'volatile', 'wchar_t', 'while', 'xor', 'xor_eq'
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
    """Calculate cognitive complexity for a C++ file."""
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
    """Calculate Halstead metrics for a C++ file."""
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
            'error': 'Usage: cpp_metrics.py <metric_type> <file_path>'
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
