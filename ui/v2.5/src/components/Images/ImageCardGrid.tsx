import React, { useMemo } from "react";
import * as GQL from "src/core/generated-graphql";
import { ImageCard } from "./ImageCard";
import {
  useCardWidth,
  useContainerDimensions,
} from "../Shared/GridCard/GridCard";
import { PatchComponent } from "src/patch";
import { useSelectHandlers } from "../List/util";

interface IImageCardGrid {
  images: GQL.SlimImageDataFragment[];
  selectedIds: Set<string>;
  zoomIndex: number;
  onSelectChange: (id: string, selected: boolean, shiftKey: boolean) => void;
  onPreview: (index: number, ev: React.MouseEvent<Element, MouseEvent>) => void;
}

const zoomWidths = [280, 340, 480, 640];

export const ImageCardGrid: React.FC<IImageCardGrid> = PatchComponent(
  "ImageCardGrid",
  ({ images, selectedIds, zoomIndex, onSelectChange, onPreview }) => {
    const [componentRef, { width: containerWidth }] = useContainerDimensions();
    const cardWidth = useCardWidth(containerWidth, zoomIndex, zoomWidths);
    const selectHandlers = useSelectHandlers(images, onSelectChange);
    const selecting = selectedIds.size > 0;
    const previewHandlers = useMemo(() => {
      if (selecting) return undefined;
      const map = new Map<string, (ev: React.MouseEvent) => void>();
      images.forEach((image, index) => {
        map.set(image.id, (ev) => onPreview(index, ev));
      });
      return map;
    }, [images, onPreview, selecting]);

    return (
      <div className="row justify-content-center" ref={componentRef}>
        {images.map((image) => (
          <ImageCard
            key={image.id}
            cardWidth={cardWidth}
            image={image}
            zoomIndex={zoomIndex}
            selecting={selecting}
            selected={selectedIds.has(image.id)}
            onSelectedChanged={selectHandlers.get(image.id)}
            onPreview={previewHandlers?.get(image.id)}
          />
        ))}
      </div>
    );
  }
);
