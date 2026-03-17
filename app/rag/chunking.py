 from typing import Iterable, List, Tuple


 def simple_chunk(text: str, max_chars: int = 4000, overlap: int = 400) -> List[str]:
     chunks: List[str] = []
     start = 0
     n = len(text)
     while start < n:
         end = min(start + max_chars, n)
         chunk = text[start:end]
         chunks.append(chunk)
         if end == n:
             break
         start = max(0, end - overlap)
     return chunks

