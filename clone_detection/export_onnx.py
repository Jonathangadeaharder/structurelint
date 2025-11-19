#!/usr/bin/env python3
"""
Export GraphCodeBERT to ONNX format with INT8 quantization.

This script exports the GraphCodeBERT model to ONNX format and applies
INT8 quantization to reduce the model size from ~500MB to ~150MB.

Usage:
    python export_onnx.py --output models/graphcodebert.onnx
    python export_onnx.py --output models/graphcodebert.onnx --quantize
"""

import argparse
import logging
from pathlib import Path

import torch
from transformers import RobertaTokenizer, RobertaModel

try:
    from onnxruntime.quantization import quantize_dynamic, QuantType
    QUANTIZATION_AVAILABLE = True
except ImportError:
    QUANTIZATION_AVAILABLE = False
    logging.warning("onnxruntime not available, quantization disabled")

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


def export_to_onnx(
    model_name: str = "microsoft/graphcodebert-base",
    output_path: str = "graphcodebert.onnx",
    opset_version: int = 14
) -> None:
    """
    Export GraphCodeBERT model to ONNX format.

    Args:
        model_name: HuggingFace model name
        output_path: Output ONNX file path
        opset_version: ONNX opset version
    """
    logger.info(f"Loading model: {model_name}")

    # Load tokenizer and model
    tokenizer = RobertaTokenizer.from_pretrained(model_name)
    model = RobertaModel.from_pretrained(model_name)
    model.eval()

    # Create dummy input for export
    dummy_text = "def hello_world():\n    print('Hello, World!')"
    inputs = tokenizer(
        dummy_text,
        return_tensors="pt",
        padding="max_length",
        max_length=512,
        truncation=True
    )

    logger.info("Exporting to ONNX...")

    # Export to ONNX
    torch.onnx.export(
        model,
        (inputs["input_ids"], inputs["attention_mask"]),
        output_path,
        export_params=True,
        opset_version=opset_version,
        do_constant_folding=True,
        input_names=["input_ids", "attention_mask"],
        output_names=["last_hidden_state", "pooler_output"],
        dynamic_axes={
            "input_ids": {0: "batch_size", 1: "sequence"},
            "attention_mask": {0: "batch_size", 1: "sequence"},
            "last_hidden_state": {0: "batch_size", 1: "sequence"},
            "pooler_output": {0: "batch_size"},
        }
    )

    logger.info(f"✓ Model exported to {output_path}")

    # Get file size
    file_size = Path(output_path).stat().st_size / (1024 * 1024)
    logger.info(f"  Model size: {file_size:.2f} MB")


def quantize_model(
    input_path: str,
    output_path: str
) -> None:
    """
    Quantize ONNX model to INT8.

    Args:
        input_path: Input ONNX model path
        output_path: Output quantized model path
    """
    if not QUANTIZATION_AVAILABLE:
        logger.error("onnxruntime not available, cannot quantize")
        return

    logger.info(f"Quantizing model: {input_path}")
    logger.info("  This may take several minutes...")

    # Quantize to INT8
    quantize_dynamic(
        model_input=input_path,
        model_output=output_path,
        weight_type=QuantType.QInt8
    )

    logger.info(f"✓ Quantized model saved to {output_path}")

    # Compare file sizes
    original_size = Path(input_path).stat().st_size / (1024 * 1024)
    quantized_size = Path(output_path).stat().st_size / (1024 * 1024)
    reduction = (1 - quantized_size / original_size) * 100

    logger.info(f"  Original size:  {original_size:.2f} MB")
    logger.info(f"  Quantized size: {quantized_size:.2f} MB")
    logger.info(f"  Reduction:      {reduction:.1f}%")


def main():
    parser = argparse.ArgumentParser(
        description="Export GraphCodeBERT to ONNX with optional quantization"
    )
    parser.add_argument(
        "--model",
        default="microsoft/graphcodebert-base",
        help="HuggingFace model name (default: microsoft/graphcodebert-base)"
    )
    parser.add_argument(
        "--output",
        default="graphcodebert.onnx",
        help="Output ONNX file path (default: graphcodebert.onnx)"
    )
    parser.add_argument(
        "--quantize",
        action="store_true",
        help="Apply INT8 quantization to reduce model size"
    )
    parser.add_argument(
        "--opset",
        type=int,
        default=14,
        help="ONNX opset version (default: 14)"
    )

    args = parser.parse_args()

    # Create output directory if needed
    output_path = Path(args.output)
    output_path.parent.mkdir(parents=True, exist_ok=True)

    # Export to ONNX
    export_to_onnx(
        model_name=args.model,
        output_path=str(output_path),
        opset_version=args.opset
    )

    # Quantize if requested
    if args.quantize:
        quantized_path = output_path.parent / f"{output_path.stem}_quantized{output_path.suffix}"
        quantize_model(
            input_path=str(output_path),
            output_path=str(quantized_path)
        )

        logger.info("\n✓ Export complete!")
        logger.info(f"  Full precision: {output_path}")
        logger.info(f"  Quantized:      {quantized_path}")
    else:
        logger.info("\n✓ Export complete!")
        logger.info(f"  Model: {output_path}")
        logger.info("\nTo quantize, run with --quantize flag")


if __name__ == "__main__":
    main()
