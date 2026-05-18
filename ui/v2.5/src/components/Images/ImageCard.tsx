import React, { MouseEvent, useEffect, useMemo, useRef } from "react";
import { Button, ButtonGroup } from "react-bootstrap";
import cx from "classnames";
import * as GQL from "src/core/generated-graphql";
import { Icon } from "src/components/Shared/Icon";
import { GalleryLink, TagLink } from "src/components/Shared/TagLink";
import { HoverPopover } from "src/components/Shared/HoverPopover";
import { PerformerPopoverButton } from "src/components/Shared/PerformerPopoverButton";
import { GridCard } from "src/components/Shared/GridCard/GridCard";
import { RatingBanner } from "src/components/Shared/RatingBanner";
import {
  faBox,
  faImages,
  faSearch,
  faTag,
} from "@fortawesome/free-solid-svg-icons";
import { imageTitle } from "src/core/files";
import { PatchComponent } from "src/patch";
import { TruncatedText } from "../Shared/TruncatedText";
import { StudioOverlay } from "../Shared/GridCard/StudioOverlay";
import { OCounterButton } from "../Shared/CountButton";

interface IImageCardProps {
  image: GQL.SlimImageDataFragment;
  cardWidth?: number;
  selecting?: boolean;
  selected?: boolean | undefined;
  zoomIndex: number;
  onSelectedChanged?: (selected: boolean, shiftKey: boolean) => void;
  onPreview?: (ev: MouseEvent) => void;
}

const ImageCardPopovers = PatchComponent(
  "ImageCard.Popovers",
  (props: IImageCardProps) => {
    function maybeRenderTagPopoverButton() {
      if (props.image.tags.length <= 0) return;

      const popoverContent = props.image.tags.map((tag) => (
        <TagLink key={tag.id} tag={tag} linkType="image" />
      ));

      return (
        <HoverPopover
          className="tag-count"
          placement="bottom"
          content={popoverContent}
        >
          <Button className="minimal">
            <Icon icon={faTag} />
            <span>{props.image.tags.length}</span>
          </Button>
        </HoverPopover>
      );
    }

    function maybeRenderPerformerPopoverButton() {
      if (props.image.performers.length <= 0) return;

      return (
        <PerformerPopoverButton
          performers={props.image.performers}
          linkType="image"
        />
      );
    }

    function maybeRenderOCounter() {
      if (props.image.o_counter) {
        return <OCounterButton value={props.image.o_counter} />;
      }
    }

    function maybeRenderGallery() {
      if (props.image.galleries.length <= 0) return;

      const popoverContent = props.image.galleries.map((gallery) => (
        <GalleryLink key={gallery.id} gallery={gallery} />
      ));

      return (
        <HoverPopover
          className="gallery-count"
          placement="bottom"
          content={popoverContent}
        >
          <Button className="minimal">
            <Icon icon={faImages} />
            <span>{props.image.galleries.length}</span>
          </Button>
        </HoverPopover>
      );
    }

    function maybeRenderOrganized() {
      if (props.image.organized) {
        return (
          <div className="organized">
            <Button className="minimal">
              <Icon icon={faBox} />
            </Button>
          </div>
        );
      }
    }

    if (
      props.image.tags.length > 0 ||
      props.image.performers.length > 0 ||
      props.image.o_counter ||
      props.image.galleries.length > 0 ||
      props.image.organized
    ) {
      return (
        <>
          <hr />
          <ButtonGroup className="card-popovers">
            {maybeRenderTagPopoverButton()}
            {maybeRenderPerformerPopoverButton()}
            {maybeRenderOCounter()}
            {maybeRenderGallery()}
            {maybeRenderOrganized()}
          </ButtonGroup>
        </>
      );
    }

    return null;
  }
);

const ImageCardDetails = PatchComponent(
  "ImageCard.Details",
  (props: IImageCardProps) => {
    return (
      <div className="image-card__details">
        <span className="image-card__date">{props.image.date}</span>
        <TruncatedText
          className="image-card__description"
          text={props.image.details}
          lineCount={3}
        />
      </div>
    );
  }
);

const ImageCardOverlays = PatchComponent(
  "ImageCard.Overlays",
  (props: IImageCardProps) => {
    const ret = useMemo(() => {
      return (
        <StudioOverlay studio={props.image.studio} disabled={props.selecting} />
      );
    }, [props.image.studio, props.selecting]);

    return ret;
  }
);

const ImageCardImage = PatchComponent(
  "ImageCard.Image",
  (props: IImageCardProps) => {
    const file = useMemo(
      () =>
        props.image.visual_files.length > 0
          ? props.image.visual_files[0]
          : undefined,
      [props.image]
    );
    const videoEl = useRef<HTMLVideoElement>(null);

    function isPortrait() {
      const width = file?.width ? file.width : 0;
      const height = file?.height ? file.height : 0;
      return height > width;
    }

    const source =
      props.image.paths.preview != ""
        ? props.image.paths.preview ?? ""
        : props.image.paths.thumbnail ?? "";
    const isVideo = source.includes("preview");

    useEffect(() => {
      if (!isVideo) return;
      const el = videoEl.current;
      if (!el) return;

      const observer = new IntersectionObserver((entries) => {
        entries.forEach((entry) => {
          if (entry.intersectionRatio > 0) {
            el.play()?.catch(() => { });
          } else {
            el.pause();
          }
        });
      });

      observer.observe(el);
      return () => observer.disconnect();
    }, [isVideo]);

    return (
      <>
        <div className={cx("image-card-preview", { portrait: isPortrait() })}>
          {isVideo ? (
            <video
              ref={videoEl}
              loop
              muted
              playsInline
              preload="none"
              disableRemotePlayback
              className="image-card-preview-image"
              src={source}
            />
          ) : (
            <img
              loading="lazy"
              decoding="async"
              className="image-card-preview-image"
              alt={props.image.title ?? ""}
              src={source}
            />
          )}
          {props.onPreview ? (
            <div className="preview-button">
              <Button onClick={props.onPreview}>
                <Icon icon={faSearch} />
              </Button>
            </div>
          ) : undefined}
        </div>
        <RatingBanner rating={props.image.rating100} />
      </>
    );
  }
);

const ImageCardComponent = (props: IImageCardProps) => {
  return (
    <GridCard
      className={`image-card zoom-${props.zoomIndex}`}
      url={`/images/${props.image.id}`}
      width={props.cardWidth}
      title={imageTitle(props.image)}
      linkClassName="image-card-link"
      image={<ImageCardImage {...props} />}
      details={<ImageCardDetails {...props} />}
      overlays={<ImageCardOverlays {...props} />}
      popovers={<ImageCardPopovers {...props} />}
      selected={props.selected}
      selecting={props.selecting}
      onSelectedChanged={props.onSelectedChanged}
    />
  );
};

export const ImageCard: React.FC<IImageCardProps> = React.memo(
  PatchComponent("ImageCard", ImageCardComponent)
);
